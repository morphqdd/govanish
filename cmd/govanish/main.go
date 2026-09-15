// Command govanish lints Go packages against a fixed, non-negotiable
// style. It has no configuration file and no suppression directives: code
// either conforms to the style or it does not.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"govanish/internal/report"
	"govanish/internal/runner"
)

const (
	exitClean    = 0
	exitFindings = 1
	exitFailure  = 2
)

// version is overridden at build time with -ldflags.
var version = "dev"

func main() {
	os.Exit(run(os.Args[1:], "", os.Stdout, os.Stderr))
}

func run(args []string, dir string, out, errOut io.Writer) int {
	flags := flag.NewFlagSet("govanish", flag.ContinueOnError)
	flags.SetOutput(errOut)

	format := flags.String("format", "text", "output format: text or json")
	showVersion := flags.Bool("version", false, "print the version and exit")

	if err := flags.Parse(args); err != nil {
		return exitFailure
	}

	if *showVersion {
		fmt.Fprintf(out, "govanish %s\n", version)

		return exitClean
	}

	render, err := renderer(*format)
	if err != nil {
		fmt.Fprintf(errOut, "govanish: %v\n", err)

		return exitFailure
	}

	patterns := flags.Args()
	if len(patterns) == 0 {
		patterns = []string{"./..."}
	}

	findings, err := runner.Run(dir, patterns)
	if err != nil {
		fmt.Fprintf(errOut, "govanish: %v\n", err)

		return exitFailure
	}

	if err := render(out, findings); err != nil {
		fmt.Fprintf(errOut, "govanish: %v\n", err)

		return exitFailure
	}

	if len(findings) > 0 {
		return exitFindings
	}

	return exitClean
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
