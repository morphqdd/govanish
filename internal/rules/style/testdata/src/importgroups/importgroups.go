// Package importgroups imports correctly.
package importgroups

import (
	"fmt"
	"strings"

	"example.com/ext"

	"importgroups/local"
)

// Use consumes every import so none is unused.
func Use() string {
	return fmt.Sprint(strings.ToUpper("x"), ext.Value, local.Value)
}
