package style

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// The three groups an import may belong to, in the order they must appear.
const (
	groupStdlib = iota
	groupExternal
	groupLocal
)

// groupNames names a group for a diagnostic.
var groupNames = map[int]string{
	groupStdlib:   "standard library",
	groupExternal: "external",
	groupLocal:    "local",
}

// ImportGroups reports imports that are not arranged as exactly three
// ordered, sorted groups: standard library, external, then local.
var ImportGroups = &analysis.Analyzer{
	Name: "importgroups",
	Doc:  "imports must form sorted groups: standard library, external, local",
	Run:  runImportGroups,
}

func runImportGroups(pass *analysis.Pass) (any, error) {
	prefix := localPrefix(pass)

	for _, file := range pass.Files {
		checkImports(pass, file, prefix)
	}

	return nil, nil
}

// localPrefix is the import path prefix that marks a package as belonging
// to the code under analysis rather than to a dependency.
func localPrefix(pass *analysis.Pass) string {
	if pass.Module != nil && pass.Module.Path != "" {
		return pass.Module.Path
	}

	return firstSegment(pass.Pkg.Path())
}

func firstSegment(path string) string {
	if index := strings.Index(path, "/"); index >= 0 {
		return path[:index]
	}

	return path
}

func checkImports(pass *analysis.Pass, file *ast.File, prefix string) {
	decls := importDecls(file)
	for _, extra := range decls[min(1, len(decls)):] {
		pass.Reportf(extra.Pos(),
			"imports must be declared in a single parenthesized block")
	}

	if len(decls) == 0 {
		return
	}

	previous := -1

	for _, group := range groupsOf(pass, decls[0]) {
		kind := checkGroup(pass, group, prefix)
		if kind <= previous {
			pass.Reportf(group[0].Pos(), "import group is out of order")
		}

		previous = max(previous, kind)
	}
}

func importDecls(file *ast.File) []*ast.GenDecl {
	var decls []*ast.GenDecl

	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if ok && gen.Tok == token.IMPORT {
			decls = append(decls, gen)
		}
	}

	return decls
}

// groupsOf splits an import block into the runs of specs that blank lines
// separate.
func groupsOf(pass *analysis.Pass, decl *ast.GenDecl) [][]*ast.ImportSpec {
	var groups [][]*ast.ImportSpec

	var current []*ast.ImportSpec

	last := 0

	for _, spec := range decl.Specs {
		imported, ok := spec.(*ast.ImportSpec)
		if !ok {
			continue
		}

		line := pass.Fset.Position(imported.Pos()).Line
		if last > 0 && line > last+1 {
			groups = append(groups, current)
			current = nil
		}

		current = append(current, imported)
		last = line
	}

	return append(groups, current)
}

// checkGroup validates one group and returns the group kind it belongs to.
func checkGroup(pass *analysis.Pass, group []*ast.ImportSpec, prefix string) int {
	kind := groupStdlib
	previousPath := ""

	for index, spec := range group {
		path := importPath(spec)
		reportSpec(pass, spec)

		if index == 0 {
			kind = groupOf(path, prefix)
		} else if got := groupOf(path, prefix); got != kind {
			pass.Reportf(spec.Pos(), "import %q belongs in the %s group, not the %s group",
				path, groupNames[got], groupNames[kind])
		}

		if previousPath != "" && path < previousPath {
			pass.Reportf(spec.Pos(), "import %q is out of order within its group", path)
		}

		previousPath = path
	}

	return kind
}

func reportSpec(pass *analysis.Pass, spec *ast.ImportSpec) {
	if spec.Name == nil {
		return
	}

	switch spec.Name.Name {
	case ".":
		pass.Reportf(spec.Pos(), "dot imports are banned")
	case "_":
		if pass.Pkg.Name() != "main" {
			pass.Reportf(spec.Pos(), "blank imports are banned outside package main")
		}
	}
}

func groupOf(path, prefix string) int {
	if prefix != "" && (path == prefix || strings.HasPrefix(path, prefix+"/")) {
		return groupLocal
	}

	if strings.Contains(firstSegment(path), ".") {
		return groupExternal
	}

	return groupStdlib
}

func importPath(spec *ast.ImportSpec) string {
	path, err := strconv.Unquote(spec.Path.Value)
	if err != nil {
		return fmt.Sprintf("%v", spec.Path.Value)
	}

	return path
}
