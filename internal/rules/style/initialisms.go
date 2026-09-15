package style

import (
	"go/ast"
	"strings"
	"unicode"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"

	"govanish/internal/astutil"
)

// The two visibilities an identifier can have.
const (
	unexported visibility = iota
	exported
)

// visibility says whether an identifier is part of a package's exported
// surface. It is a named type because govanish bans boolean parameters.
type visibility int

// Initialisms reports identifiers that spell a known initialism in mixed
// case, such as ParseUrl instead of ParseURL.
func Initialisms() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "initialisms",
		Doc:      "initialisms such as URL and ID must be spelled in full capitals",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runInitialisms,
	}
}

// initialisms are the words Go spells in full capitals wherever they
// appear in an identifier.
func initialisms() map[string]bool {
	return map[string]bool{
		"ACL": true, "API": true, "ASCII": true, "CPU": true, "CSS": true,
		"DNS": true, "EOF": true, "GUID": true, "HTML": true, "HTTP": true,
		"HTTPS": true, "ID": true, "IP": true, "JSON": true, "LHS": true,
		"QPS": true, "RAM": true, "RHS": true, "RPC": true, "SLA": true,
		"SMTP": true, "SQL": true, "SSH": true, "TCP": true, "TLS": true,
		"TTL": true, "UDP": true, "UI": true, "UID": true, "UUID": true,
		"URI": true, "URL": true, "UTF8": true, "VM": true, "XML": true,
		"XMPP": true, "XSRF": true, "XSS": true,
	}
}

func visibilityOf(ident *ast.Ident) visibility {
	if ident.IsExported() {
		return exported
	}

	return unexported
}

func runInitialisms(pass *analysis.Pass) (any, error) {
	insp, err := astutil.Inspector(pass)
	if err != nil {
		return nil, err
	}

	packageNames := packageIdents(pass)

	insp.Preorder([]ast.Node{(*ast.Ident)(nil)}, func(node ast.Node) {
		ident, ok := node.(*ast.Ident)
		if !ok || packageNames[ident] || !isDeclaration(pass, ident) {
			return
		}

		fixed := correctInitialisms(ident.Name, visibilityOf(ident))
		if fixed == ident.Name {
			return
		}

		pass.Reportf(ident.Pos(), "%s should be %s", ident.Name, fixed)
	})

	return nil, nil
}

// packageIdents collects the identifiers naming the package itself,
// which Go spells in lowercase whatever they contain.
func packageIdents(pass *analysis.Pass) map[*ast.Ident]bool {
	names := make(map[*ast.Ident]bool)

	for _, file := range pass.Files {
		names[file.Name] = true
	}

	return names
}

// correctInitialisms returns the identifier with every word that names an
// initialism capitalized in full. The first word of an unexported
// identifier is left alone: Go writes urlParser, not URLParser.
func correctInitialisms(name string, kind visibility) string {
	words := splitWords(name)

	for index, word := range words {
		if index == 0 && kind == unexported {
			continue
		}

		upper := strings.ToUpper(word)
		if initialisms()[upper] {
			words[index] = upper
		}
	}

	return strings.Join(words, "")
}

// splitWords breaks a camel case identifier at each capital letter. A run
// of capitals stays together, so "ParseURLNow" splits as Parse, URLNow.
func splitWords(name string) []string {
	var words []string

	start := 0

	for index, symbol := range name {
		if index == 0 || !unicode.IsUpper(symbol) {
			continue
		}

		words = append(words, name[start:index])
		start = index
	}

	return append(words, name[start:])
}
