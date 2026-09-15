// Package size holds the rules that cap how large a unit of code may
// grow: lines, parameters, results, fields, branches and nesting.
//
// Every limit in this package is a constant. A limit that can be raised
// from a config file is not a limit.
package size
