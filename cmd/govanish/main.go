// Command govanish lints Go packages against a fixed, non-negotiable
// style. It has no configuration file and no suppression directives: code
// either conforms to the style or it does not.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/morphqdd/govanish/internal/report"
	"github.com/morphqdd/govanish/internal/runner"
)

const (
	exitClean    = 0
	exitFindings = 1
	exitFailure  = 2
)

// version names this build. It is a constant because govanish bans
// package-level variables, which rules out the -ldflags -X trick.
const version = "dev"

// options is everything one lint run needs. It exists because govanish
// forbids functions of five parameters, and it was right to.
type options struct {
	format   string
	patterns []string
	dir      runner.Dir
	out      io.Writer
	errOut   io.Writer
}

func main() {
	os.Exit(run(os.Args[1:], "", os.Stdout, os.Stderr))
}

func run(args []string, dir runner.Dir, out, errOut io.Writer) int {
	flags := flag.NewFlagSet("govanish", flag.ContinueOnError)
	flags.SetOutput(errOut)

	format := flags.String("format", "text", "output format: text or json")
	showVersion := flags.Bool("version", false, "print the version and exit")

	if err := flags.Parse(args); err != nil {
		return exitFailure
	}

	if *showVersion {
		if _, err := fmt.Fprintf(out, "govanish %s\n", version); err != nil {
			return exitFailure
		}

		return exitClean
	}

	return lint(options{
		format:   *format,
		patterns: patternsOf(flags),
		dir:      dir,
		out:      out,
		errOut:   errOut,
	})
}

func lint(opts options) int {
	render, err := renderer(opts.format)
	if err != nil {
		return fail(opts.errOut, err)
	}

	findings, err := runner.Run(opts.dir, opts.patterns)
	if err != nil {
		return fail(opts.errOut, err)
	}

	if renderErr := render(opts.out, findings); renderErr != nil {
		return fail(opts.errOut, renderErr)
	}

	if len(findings) > 0 {
		return exitFindings
	}

	return exitClean
}

// patternsOf returns the package patterns given on the command line, or
// the whole module when none were.
func patternsOf(flags *flag.FlagSet) []string {
	patterns := flags.Args()
	if len(patterns) == 0 {
		return []string{"./..."}
	}

	return patterns
}

// fail reports err and yields the failure exit code. A failure to write
// the message itself changes nothing: the command is already failing.
func fail(errOut io.Writer, err error) int {
	if _, writeErr := fmt.Fprintf(errOut, "govanish: %v\n", err); writeErr != nil {
		return exitFailure
	}

	return exitFailure
}

func renderer(format string) (func(io.Writer, []report.Finding) error, error) {
	switch format {
	case "text":
		return report.Text, nil
	case "json":
		return report.JSON, nil
	default:
		return nil, fmt.Errorf("unknown format %q: want text or json", format)
	}
}
