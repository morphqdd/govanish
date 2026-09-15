package style

import (
	"go/ast"
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
func groupNames() map[int]string {
	return map[int]string{
		groupStdlib:   "standard library",
		groupExternal: "external",
		groupLocal:    "local",
	}
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

// groupOf classifies an import path. A first segment containing a dot is
// a domain name, which means somebody else's code.
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
		return spec.Path.Value
	}

	return path
}
