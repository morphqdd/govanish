// Package report renders findings in the formats govanish supports.
package report

import (
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"slices"
)

// Finding is one rule violation at one source position.
type Finding struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Col     int    `json:"col"`
	EndLine int    `json:"end_line"`
	EndCol  int    `json:"end_col"`
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

// Sort orders findings by file, line, column, then rule identifier, so
// that output is stable across runs regardless of analysis order.
func Sort(findings []Finding) {
	slices.SortFunc(findings, func(left, right Finding) int {
		return cmp.Or(
			cmp.Compare(left.File, right.File),
			cmp.Compare(left.Line, right.Line),
			cmp.Compare(left.Col, right.Col),
			cmp.Compare(left.Rule, right.Rule),
		)
	})
}

// Text writes findings one per line as file:line:col: [rule] message.
func Text(writer io.Writer, findings []Finding) error {
	for _, finding := range findings {
		_, err := fmt.Fprintf(writer, "%s:%d:%d: [%s] %s\n",
			finding.File, finding.Line, finding.Col, finding.Rule, finding.Message)
		if err != nil {
			return fmt.Errorf("write finding: %w", err)
		}
	}

	return nil
}

// JSON writes findings as a flat array, empty rather than null when there
// are none.
func JSON(writer io.Writer, findings []Finding) error {
	if findings == nil {
		findings = []Finding{}
	}

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(findings); err != nil {
		return fmt.Errorf("encode findings: %w", err)
	}

	return nil
}
