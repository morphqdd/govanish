// Package good conforms to every rule, so that the command's clean-run
// test stays clean as rules are added.
package good

// Answer is documented, because every exported thing must be.
func Answer() int {
	return 42
}
