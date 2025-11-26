// Package analyzer provides the core analysis engine for detecting
// constant-time violations in disassembled code.
package analyzer

import (
	"github.com/trailofbits/ml-dsa/test_ct/arch"
	"github.com/trailofbits/ml-dsa/test_ct/disasm"
)

// Config holds the configuration for analysis.
type Config struct {
	// Architecture is the target architecture analyzer.
	Architecture arch.Architecture

	// Functions are the disassembled functions to analyze.
	Functions []disasm.Function

	// IncludeWarnings includes warning-level violations (e.g., conditional branches).
	IncludeWarnings bool

	// Verbose enables verbose output during analysis.
	Verbose bool
}

// Violation represents a detected constant-time violation.
type Violation struct {
	// Function is the name of the function containing the violation.
	Function string

	// ShortFunction is the short name of the function.
	ShortFunction string

	// File is the source file path.
	File string

	// Address is the instruction address.
	Address uint64

	// Instruction is the violating instruction.
	Instruction string

	// Mnemonic is the instruction mnemonic.
	Mnemonic string

	// Reason explains why this instruction is dangerous.
	Reason string

	// Severity is either "error" or "warning".
	Severity string
}

// Report contains the results of an analysis.
type Report struct {
	// Architecture is the target architecture name.
	Architecture string

	// GOOS is the target operating system.
	GOOS string

	// GOARCH is the target architecture.
	GOARCH string

	// TotalFunctions is the number of functions analyzed.
	TotalFunctions int

	// TotalInstructions is the total number of instructions analyzed.
	TotalInstructions int

	// Violations contains all detected violations.
	Violations []Violation

	// ErrorCount is the number of error-level violations.
	ErrorCount int

	// WarningCount is the number of warning-level violations.
	WarningCount int

	// Passed indicates whether the analysis passed (no errors).
	Passed bool
}

// Analyze performs constant-time analysis on the given functions.
func Analyze(cfg Config) *Report {
	report := &Report{
		Architecture:   cfg.Architecture.Name(),
		GOARCH:         cfg.Architecture.GOARCH(),
		TotalFunctions: len(cfg.Functions),
	}

	for _, fn := range cfg.Functions {
		report.TotalInstructions += len(fn.Instructions)

		// Analyze instructions using the architecture
		archViolations := cfg.Architecture.AnalyzeInstructions(fn.Instructions)

		for _, v := range archViolations {
			// Skip warnings if not included
			if v.Severity == "warning" && !cfg.IncludeWarnings {
				continue
			}

			var instr arch.Instruction
			if v.InstructionIndex >= 0 && v.InstructionIndex < len(fn.Instructions) {
				instr = fn.Instructions[v.InstructionIndex]
			}

			violation := Violation{
				Function:      fn.Name,
				ShortFunction: fn.ShortName,
				File:          fn.File,
				Address:       instr.Address,
				Instruction:   v.Instruction,
				Mnemonic:      instr.Mnemonic,
				Reason:        v.Reason,
				Severity:      v.Severity,
			}

			report.Violations = append(report.Violations, violation)

			if v.Severity == "error" {
				report.ErrorCount++
			} else {
				report.WarningCount++
			}
		}
	}

	// Analysis passes if there are no errors
	report.Passed = report.ErrorCount == 0

	return report
}

// AnalyzeSingle analyzes a single function and returns violations.
func AnalyzeSingle(architecture arch.Architecture, fn disasm.Function, includeWarnings bool) []Violation {
	cfg := Config{
		Architecture:    architecture,
		Functions:       []disasm.Function{fn},
		IncludeWarnings: includeWarnings,
	}
	report := Analyze(cfg)
	return report.Violations
}
