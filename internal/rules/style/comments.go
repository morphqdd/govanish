package style

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Comments reports every comment that is not a doc comment on a
// declaration. A comment explaining code inside a function is a sign the
// code needs a name, not a note.
func Comments() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "comments",
		Doc:  "comments are allowed only as doc comments on declarations",
		Run:  runComments,
	}
}

// bannedMarkers are the comment markers that record work left undone.
// Undone work belongs in an issue tracker, where it is visible.
func bannedMarkers() []string {
	return []string{"TODO", "FIXME", "XXX", "HACK"}
}

func runComments(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		checkFileComments(pass, file)
	}

	return nil, nil
}

func checkFileComments(pass *analysis.Pass, file *ast.File) {
	allowed := docComments(file)

	for _, group := range file.Comments {
		if isDirective(group) {
			continue
		}

		if !allowed[group] {
			pass.Reportf(group.Pos(), "comment is not a doc comment")

			continue
		}

		reportMarkers(pass, group)
	}
}

// docComments collects every comment group the syntax tree attaches to a
// declaration as documentation.
func docComments(file *ast.File) map[*ast.CommentGroup]bool {
	allowed := make(map[*ast.CommentGroup]bool)

	add := func(group *ast.CommentGroup) {
		if group != nil {
			allowed[group] = true
		}
	}

	add(file.Doc)

	ast.Inspect(file, func(node ast.Node) bool {
		add(docOf(node))

		return true
	})

	return allowed
}

func docOf(node ast.Node) *ast.CommentGroup {
	switch typed := node.(type) {
	case *ast.GenDecl:
		return typed.Doc
	case *ast.FuncDecl:
		return typed.Doc
	case *ast.TypeSpec:
		return typed.Doc
	case *ast.ValueSpec:
		return typed.Doc
	case *ast.ImportSpec:
		return typed.Doc
	case *ast.Field:
		return typed.Doc
	default:
		return nil
	}
}

// isDirective reports whether the group is a compiler directive such as
// //go:build or //go:noinline, which carry meaning rather than prose.
func isDirective(group *ast.CommentGroup) bool {
	for _, comment := range group.List {
		if !strings.HasPrefix(comment.Text, "//go:") &&
			!strings.HasPrefix(comment.Text, "//line ") &&
			!strings.HasPrefix(comment.Text, "// +build") {
			return false
		}
	}

	return true
}

func reportMarkers(pass *analysis.Pass, group *ast.CommentGroup) {
	for _, comment := range group.List {
		for _, marker := range bannedMarkers() {
			if !strings.Contains(comment.Text, marker) {
				continue
			}

			pass.Reportf(comment.Pos(), "TODO and FIXME comments are banned")

			return
		}
	}
}
