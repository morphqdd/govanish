// Package importbad imports incorrectly.
package importbad

import (
	"example.com/ext"

	"strings" // want `import group is out of order`
	"fmt"     // want `import "fmt" is out of order within its group`
)

import "sort" // want `imports must be declared in a single parenthesized block`

// Use consumes every import so none is unused.
func Use() string {
	sort.Strings(nil)

	return fmt.Sprint(strings.ToUpper("x"), ext.Value)
}
