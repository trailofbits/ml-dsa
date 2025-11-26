package analyzer

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// ReportFormat specifies the output format for reports.
type ReportFormat string

const (
	// FormatText outputs a human-readable text report.
	FormatText ReportFormat = "text"

	// FormatJSON outputs a JSON report.
	FormatJSON ReportFormat = "json"

	// FormatGitHubActions outputs annotations for GitHub Actions.
	FormatGitHubActions ReportFormat = "github"
)

// WriteReport writes a report to the given writer in the specified format.
func WriteReport(w io.Writer, report *Report, format ReportFormat) error {
	switch format {
	case FormatJSON:
		return writeJSONReport(w, report)
	case FormatGitHubActions:
		return writeGitHubActionsReport(w, report)
	default:
		return writeTextReport(w, report)
	}
}

// writeTextReport writes a human-readable text report.
func writeTextReport(w io.Writer, report *Report) error {
	fmt.Fprintf(w, "Constant-Time Analysis Report\n")
	fmt.Fprintf(w, "==============================\n\n")
	fmt.Fprintf(w, "Architecture: %s (%s)\n", report.Architecture, report.GOARCH)
	fmt.Fprintf(w, "Functions analyzed: %d\n", report.TotalFunctions)
	fmt.Fprintf(w, "Instructions analyzed: %d\n\n", report.TotalInstructions)

	if len(report.Violations) == 0 {
		fmt.Fprintf(w, "No violations found.\n\n")
	} else {
		// Group violations by function
		byFunction := make(map[string][]Violation)
		for _, v := range report.Violations {
			byFunction[v.Function] = append(byFunction[v.Function], v)
		}

		for fn, violations := range byFunction {
			fmt.Fprintf(w, "Function: %s\n", fn)
			if violations[0].File != "" {
				fmt.Fprintf(w, "  File: %s\n", violations[0].File)
			}
			fmt.Fprintf(w, "\n")

			for _, v := range violations {
				severityStr := strings.ToUpper(v.Severity)
				fmt.Fprintf(w, "  [%s] 0x%x: %s\n", severityStr, v.Address, v.Instruction)
				fmt.Fprintf(w, "         %s\n\n", v.Reason)
			}
		}
	}

	fmt.Fprintf(w, "Summary\n")
	fmt.Fprintf(w, "-------\n")
	fmt.Fprintf(w, "Errors:   %d\n", report.ErrorCount)
	fmt.Fprintf(w, "Warnings: %d\n", report.WarningCount)

	if report.Passed {
		fmt.Fprintf(w, "\nPASSED\n")
	} else {
		fmt.Fprintf(w, "\nFAILED\n")
	}

	return nil
}

// JSONReport is the JSON-serializable report structure.
type JSONReport struct {
	Architecture      string          `json:"architecture"`
	GOARCH            string          `json:"goarch"`
	GOOS              string          `json:"goos,omitempty"`
	TotalFunctions    int             `json:"total_functions"`
	TotalInstructions int             `json:"total_instructions"`
	ErrorCount        int             `json:"error_count"`
	WarningCount      int             `json:"warning_count"`
	Passed            bool            `json:"passed"`
	Violations        []JSONViolation `json:"violations"`
}

// JSONViolation is the JSON-serializable violation structure.
type JSONViolation struct {
	Function    string `json:"function"`
	File        string `json:"file,omitempty"`
	Address     string `json:"address"`
	Instruction string `json:"instruction"`
	Mnemonic    string `json:"mnemonic"`
	Reason      string `json:"reason"`
	Severity    string `json:"severity"`
}

// writeJSONReport writes a JSON report.
func writeJSONReport(w io.Writer, report *Report) error {
	jsonReport := JSONReport{
		Architecture:      report.Architecture,
		GOARCH:            report.GOARCH,
		GOOS:              report.GOOS,
		TotalFunctions:    report.TotalFunctions,
		TotalInstructions: report.TotalInstructions,
		ErrorCount:        report.ErrorCount,
		WarningCount:      report.WarningCount,
		Passed:            report.Passed,
		Violations:        make([]JSONViolation, len(report.Violations)),
	}

	for i, v := range report.Violations {
		jsonReport.Violations[i] = JSONViolation{
			Function:    v.Function,
			File:        v.File,
			Address:     fmt.Sprintf("0x%x", v.Address),
			Instruction: v.Instruction,
			Mnemonic:    v.Mnemonic,
			Reason:      v.Reason,
			Severity:    v.Severity,
		}
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(jsonReport)
}

// writeGitHubActionsReport writes GitHub Actions workflow commands.
func writeGitHubActionsReport(w io.Writer, report *Report) error {
	for _, v := range report.Violations {
		// GitHub Actions annotation format:
		// ::error file={name},line={line},endLine={endLine},title={title}::{message}
		// ::warning file={name},line={line},endLine={endLine},title={title}::{message}

		level := "error"
		if v.Severity == "warning" {
			level = "warning"
		}

		title := fmt.Sprintf("Constant-time violation: %s", v.Mnemonic)
		message := fmt.Sprintf("%s in %s: %s", v.Instruction, v.ShortFunction, v.Reason)

		if v.File != "" {
			fmt.Fprintf(w, "::%s file=%s,title=%s::%s\n", level, v.File, title, message)
		} else {
			fmt.Fprintf(w, "::%s title=%s::%s\n", level, title, message)
		}
	}

	// Write summary
	if report.Passed {
		fmt.Fprintf(w, "::notice::Constant-time analysis passed. %d functions, %d instructions analyzed.\n",
			report.TotalFunctions, report.TotalInstructions)
	} else {
		fmt.Fprintf(w, "::error::Constant-time analysis failed. %d error(s) found.\n", report.ErrorCount)
	}

	return nil
}
