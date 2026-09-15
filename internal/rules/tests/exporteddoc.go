package tests

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// docTarget is one declaration to check, gathered into a struct because
// govanish bans functions of five parameters.
type docTarget struct {
	node ast.Node
	kind string
	name *ast.Ident
	doc  *ast.CommentGroup
}

// ExportedDoc reports exported declarations without a doc comment, and
// doc comments that do not begin with the name they document.
func ExportedDoc() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "exporteddoc",
		Doc:  "exported declarations must have a doc comment beginning with their name",
		Run:  runExportedDoc,
	}
}

func runExportedDoc(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if isTestFile(pass.Fset.Position(file.Pos()).Filename) {
			continue
		}

		for _, decl := range file.Decls {
			checkExportedDoc(pass, decl)
		}
	}

	return nil, nil
}

func checkExportedDoc(pass *analysis.Pass, decl ast.Decl) {
	switch typed := decl.(type) {
	case *ast.FuncDecl:
		checkFuncDoc(pass, typed)
	case *ast.GenDecl:
		checkGenDoc(pass, typed)
	}
}

func checkFuncDoc(pass *analysis.Pass, decl *ast.FuncDecl) {
	if decl.Recv != nil {
		reportDoc(pass, docTarget{
			node: decl, kind: "method", name: decl.Name, doc: decl.Doc,
		})

		return
	}

	reportDoc(pass, docTarget{
		node: decl, kind: "function", name: decl.Name, doc: decl.Doc,
	})
}

// checkGenDoc checks a general declaration. A grouped declaration carries
// its documentation on each spec; a single one carries it on the group.
func checkGenDoc(pass *analysis.Pass, decl *ast.GenDecl) {
	for _, spec := range decl.Specs {
		typed, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}

		doc := typed.Doc
		if doc == nil {
			doc = decl.Doc
		}

		reportDoc(pass, docTarget{
			node: decl, kind: "type", name: typed.Name, doc: doc,
		})
	}
}

func reportDoc(pass *analysis.Pass, target docTarget) {
	if !target.name.IsExported() {
		return
	}

	if target.doc == nil {
		pass.Reportf(target.node.Pos(), "exported %s %s has no doc comment",
			target.kind, target.name.Name)

		return
	}

	if !strings.HasPrefix(strings.TrimSpace(target.doc.Text()), target.name.Name) {
		pass.Reportf(target.node.Pos(), "doc comment for %s must begin with its name",
			target.name.Name)
	}
}
