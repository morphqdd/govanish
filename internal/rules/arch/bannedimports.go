package arch

import (
	"go/token"
	"strconv"

	"golang.org/x/tools/go/analysis"
)

// BannedImports reports imports of packages govanish does not allow.
func BannedImports() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "bannedimports",
		Doc:  "some standard library packages are banned outright",
		Run:  runBannedImports,
	}
}

// bannedImports() maps a forbidden import path to the reason it is
// forbidden, or to the empty string when the ban needs no elaboration.
func bannedImports() map[string]string {
	return map[string]string{
		"log":       "use log/slog",
		"math/rand": "use crypto/rand",
		"reflect":   "",
		"unsafe":    "",
	}
}

func runBannedImports(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		for _, spec := range file.Imports {
			path, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				continue
			}

			reportBanned(pass, spec.Pos(), path)
		}
	}

	return nil, nil
}

func reportBanned(pass *analysis.Pass, position token.Pos, path string) {
	reason, banned := bannedImports()[path]
	if !banned {
		return
	}

	if reason == "" {
		pass.Reportf(position, "import %q is banned", path)

		return
	}

	pass.Reportf(position, "import %q is banned; %s", path, reason)
}
