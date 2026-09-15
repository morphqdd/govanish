// Package ctxorigin exercises where a context may be created.
package ctxorigin

import "context"

// Roots a context that should have been passed in.
func Roots() context.Context {
	return context.Background() // want `context.Background may only be called in main or a test`
}

// Todos is no better.
func Todos() context.Context {
	return context.TODO() // want `context.TODO may only be called in main or a test`
}

// Derives from a context it was given, which is correct.
func Derives(ctx context.Context) context.Context {
	return ctx
}
