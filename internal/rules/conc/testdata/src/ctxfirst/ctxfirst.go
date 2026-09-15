// Package ctxfirst exercises where a context goes.
package ctxfirst

import "context"

// Good takes the context first and calls it ctx.
func Good(ctx context.Context, name string) {}

// Late takes it second.
func Late(name string, ctx context.Context) {} // want `context.Context must be the first parameter`

// Misnamed calls it something else.
func Misnamed(c context.Context) {} // want `context parameter must be named ctx, not c`

// Holder stores a context, which outlives the call it belongs to.
type Holder struct {
	ctx context.Context // want `context.Context must not be stored in a struct field`
}

// None takes no context at all.
func None(name string) {}
