// Package wholeprogram holds the checks that no single package can make
// about itself. An analysis.Analyzer sees a package and its dependencies,
// never its importers, so a question like "does anyone use this?" has to
// be asked of the whole program at once.
package wholeprogram

import (
	"go/types"

	"golang.org/x/tools/go/packages"

	"govanish/internal/report"
)

// minExportID is the rule identifier printed with these findings.
const minExportID = "min-export"

// MinExport reports exported identifiers that nothing outside their own
// package uses. An export nobody needs is a promise nobody asked for.
//
// A package named main is skipped: its surface is the program itself.
// So is a test package, whose exports exist for the test binary.
func MinExport(pkgs []*packages.Package) []report.Finding {
	used := usedObjects(pkgs)
	expandSignatures(pkgs, used)
	findings := make([]report.Finding, 0)

	for _, pkg := range pkgs {
		if skipPackage(pkg) {
			continue
		}

		findings = append(findings, orphanExports(pkg, used)...)
	}

	return findings
}

func skipPackage(pkg *packages.Package) bool {
	return pkg.Name == "main" || pkg.Name == "" || len(pkg.Syntax) == 0
}

// usedObjects collects every object each package references from a
// package other than its own.
func usedObjects(pkgs []*packages.Package) map[types.Object]bool {
	used := make(map[types.Object]bool)

	for _, pkg := range pkgs {
		if pkg.TypesInfo == nil {
			continue
		}

		for _, object := range pkg.TypesInfo.Uses {
			if object == nil || object.Pkg() == nil {
				continue
			}

			if object.Pkg() != pkg.Types {
				used[object] = true
			}
		}
	}

	return used
}

// expandSignatures marks the types named in the signature of anything
// used from outside. A type returned by an exported function is part of
// the surface even when no caller writes its name.
func expandSignatures(pkgs []*packages.Package, used map[types.Object]bool) {
	for _, pkg := range pkgs {
		if pkg.Types == nil {
			continue
		}

		scope := pkg.Types.Scope()

		for _, name := range scope.Names() {
			object := scope.Lookup(name)
			if used[object] {
				markNamedTypes(object.Type(), used)
			}
		}
	}
}

func markNamedTypes(typ types.Type, used map[types.Object]bool) {
	if named, isNamed := typ.(*types.Named); isNamed {
		markMethods(named, used)

		return
	}

	signature, ok := typ.(*types.Signature)
	if !ok {
		return
	}

	for _, tuple := range []*types.Tuple{signature.Params(), signature.Results()} {
		for index := range tuple.Len() {
			markNamed(tuple.At(index).Type(), used)
		}
	}
}

// markMethods marks everything the methods of a used type mention. A
// type reachable only as a method result is still part of the surface.
func markMethods(named *types.Named, used map[types.Object]bool) {
	for index := range named.NumMethods() {
		markNamedTypes(named.Method(index).Type(), used)
	}
}

func markNamed(typ types.Type, used map[types.Object]bool) {
	switch typed := typ.(type) {
	case *types.Named:
		used[typed.Obj()] = true
	case *types.Slice:
		markNamed(typed.Elem(), used)
	case *types.Pointer:
		markNamed(typed.Elem(), used)
	}
}

func orphanExports(pkg *packages.Package, used map[types.Object]bool) []report.Finding {
	var findings []report.Finding

	scope := pkg.Types.Scope()

	for _, name := range scope.Names() {
		object := scope.Lookup(name)
		if !object.Exported() || used[object] {
			continue
		}

		findings = append(findings, finding(pkg, object))
	}

	return findings
}

func finding(pkg *packages.Package, object types.Object) report.Finding {
	position := pkg.Fset.Position(object.Pos())

	return report.Finding{
		File:    position.Filename,
		Line:    position.Line,
		Col:     position.Column,
		EndLine: position.Line,
		EndCol:  position.Column,
		Rule:    minExportID,
		Message: "exported " + kindOf(object) + " " + object.Name() +
			" is not used outside its package",
	}
}

func kindOf(object types.Object) string {
	switch object.(type) {
	case *types.Func:
		return "function"
	case *types.TypeName:
		return "type"
	case *types.Const:
		return "constant"
	default:
		return "identifier"
	}
}
