// Package bannedimports exercises the import blocklist.
package bannedimports

import (
	"log"     // want `import "log" is banned; use log/slog`
	"reflect" // want `import "reflect" is banned`
	"strings"
)

// Use consumes the imports.
func Use() string {
	log.Print("x")
	_ = reflect.TypeOf(1)

	return strings.ToUpper("x")
}
