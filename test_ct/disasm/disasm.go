// Package disasm provides functionality to disassemble Go binaries and extract
// function assembly for constant-time analysis.
package disasm

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/trailofbits/ml-dsa/test_ct/arch"
)

// Config holds the configuration for disassembly.
type Config struct {
	// BinaryPath is the path to the compiled binary.
	BinaryPath string

	// FunctionPattern is a regex pattern to filter functions.
	// If empty, all functions are returned.
	FunctionPattern string
}

// Function represents a disassembled function.
type Function struct {
	// Name is the full name of the function (e.g., "github.com/foo/bar.MyFunc").
	Name string

	// ShortName is the short name without package path (e.g., "bar.MyFunc").
	ShortName string

	// File is the source file where the function is defined.
	File string

	// Line is the starting line number in the source file.
	Line int

	// Instructions contains the disassembled instructions.
	Instructions []arch.Instruction
}

// Disassemble runs go tool objdump on a binary and parses the output.
func Disassemble(cfg Config) ([]Function, error) {
	if cfg.BinaryPath == "" {
		return nil, fmt.Errorf("BinaryPath is required")
	}

	// Build objdump command
	args := []string{"tool", "objdump"}
	if cfg.FunctionPattern != "" {
		args = append(args, "-s", cfg.FunctionPattern)
	}
	args = append(args, cfg.BinaryPath)

	cmd := exec.Command("go", args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("go tool objdump failed: %w\nStderr: %s", err, string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("go tool objdump failed: %w", err)
	}

	return parseObjdumpOutput(string(output))
}

// parseObjdumpOutput parses the output of go tool objdump.
// The format is:
//
//	TEXT symbol_name(SB) source_file
//	  source_file:line_number     address: instruction
//	  ...
func parseObjdumpOutput(output string) ([]Function, error) {
	var functions []Function
	var currentFunc *Function

	// Regex to match function header: TEXT symbol_name(SB) source_file
	funcHeaderRe := regexp.MustCompile(`^TEXT\s+([^\s(]+)\(SB\)\s+(.*)$`)

	// Regex to match instruction line: source_file:line  address  hex_bytes  instruction
	// Example: main.go:5		0x46fb8b		48f7f3			DIVQ BX
	// The format uses tabs as separators and has hex bytes between address and instruction
	instructionRe := regexp.MustCompile(`^\s*([^:]+):(\d+)\s+0x([0-9a-fA-F]+)\s+[0-9a-fA-F]+\s+(.+?)\s*$`)

	// Alternative instruction format without source info
	// Example:   0x1234          MOVQ AX, BX
	instructionNoSourceRe := regexp.MustCompile(`^\s+0x([0-9a-fA-F]+)\s+(.+)$`)

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()

		// Check for function header
		if matches := funcHeaderRe.FindStringSubmatch(line); matches != nil {
			// Save previous function if any
			if currentFunc != nil && len(currentFunc.Instructions) > 0 {
				functions = append(functions, *currentFunc)
			}

			fullName := matches[1]
			sourceFile := strings.TrimSpace(matches[2])

			currentFunc = &Function{
				Name:      fullName,
				ShortName: shortFunctionName(fullName),
				File:      sourceFile,
			}
			continue
		}

		// Skip if we're not in a function
		if currentFunc == nil {
			continue
		}

		// Try to match instruction with source info
		if matches := instructionRe.FindStringSubmatch(line); matches != nil {
			file := matches[1]
			lineNum, _ := strconv.Atoi(matches[2])
			addrStr := matches[3]
			instr := strings.TrimSpace(matches[4])

			addr, _ := strconv.ParseUint(addrStr, 16, 64)

			// Update function file/line if not set
			if currentFunc.File == "" {
				currentFunc.File = file
			}
			if currentFunc.Line == 0 {
				currentFunc.Line = lineNum
			}

			// Parse mnemonic and operands
			mnemonic, operands := parseInstruction(instr)

			currentFunc.Instructions = append(currentFunc.Instructions, arch.Instruction{
				Address:  addr,
				Mnemonic: mnemonic,
				Operands: operands,
				Raw:      instr,
			})
			continue
		}

		// Try to match instruction without source info
		if matches := instructionNoSourceRe.FindStringSubmatch(line); matches != nil {
			addrStr := matches[1]
			instr := strings.TrimSpace(matches[2])

			addr, _ := strconv.ParseUint(addrStr, 16, 64)
			mnemonic, operands := parseInstruction(instr)

			currentFunc.Instructions = append(currentFunc.Instructions, arch.Instruction{
				Address:  addr,
				Mnemonic: mnemonic,
				Operands: operands,
				Raw:      instr,
			})
		}
	}

	// Don't forget the last function
	if currentFunc != nil && len(currentFunc.Instructions) > 0 {
		functions = append(functions, *currentFunc)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading objdump output: %w", err)
	}

	return functions, nil
}

// parseInstruction splits an instruction into mnemonic and operands.
func parseInstruction(instr string) (mnemonic, operands string) {
	// Handle instructions with comments (e.g., "MOVQ AX, BX // some comment")
	if idx := strings.Index(instr, "//"); idx != -1 {
		instr = strings.TrimSpace(instr[:idx])
	}

	// Split on first whitespace
	parts := strings.SplitN(instr, " ", 2)
	mnemonic = strings.TrimSpace(parts[0])
	if len(parts) > 1 {
		operands = strings.TrimSpace(parts[1])
	}

	// Remove any suffixes that might be part of the mnemonic (e.g., "MOVQ" -> "MOVQ")
	// But keep condition codes separate (e.g., "B.EQ" stays as is)

	return mnemonic, operands
}

// shortFunctionName extracts the short name from a full function path.
// E.g., "github.com/foo/bar.MyFunc" -> "bar.MyFunc"
func shortFunctionName(fullName string) string {
	// Find the last slash
	lastSlash := strings.LastIndex(fullName, "/")
	if lastSlash == -1 {
		return fullName
	}
	return fullName[lastSlash+1:]
}

// FilterFunctions filters a slice of functions by a regex pattern.
func FilterFunctions(funcs []Function, pattern string) ([]Function, error) {
	if pattern == "" {
		return funcs, nil
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid function pattern: %w", err)
	}

	var filtered []Function
	for _, f := range funcs {
		if re.MatchString(f.Name) || re.MatchString(f.ShortName) {
			filtered = append(filtered, f)
		}
	}

	return filtered, nil
}
