// Package good conforms to every rule, so that the command's clean-run
// test stays clean as rules are added. It exports nothing: min-export
// reports an export that no other package consumes, and this fixture has
// no consumers.
package good

func answer() int {
	return 42
}
