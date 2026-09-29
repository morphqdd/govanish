package wholeprogram

import "go/types"

// objectKey names a package-level object by its package path and name.
// Loading with Tests set type-checks a package with in-package tests
// twice, and each variant has its own types.Object for one declaration,
// so a use recorded against one variant must also count for the other.
type objectKey struct {
	path string
	name string
}

// usage is the set of package-level objects used from outside.
type usage map[objectKey]bool

// keyOf returns the key of a package-level object. Methods, fields and
// locals have no key: only package scope can hold an orphaned export.
func keyOf(object types.Object) (objectKey, bool) {
	pkg := object.Pkg()
	if pkg == nil || object.Parent() != pkg.Scope() {
		return objectKey{}, false
	}

	return objectKey{path: pkg.Path(), name: object.Name()}, true
}

func (us usage) mark(object types.Object) {
	if key, ok := keyOf(object); ok {
		us[key] = true
	}
}

func (us usage) has(object types.Object) bool {
	key, ok := keyOf(object)

	return ok && us[key]
}
