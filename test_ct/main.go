// test_ct is a constant-time verification tool for Go cryptographic code.
//
// It cross-compiles Go modules for target platforms, disassembles the resulting
// binaries, and analyzes the assembly for instructions that could leak timing
// information (such as variable-time division, conditional branches on secret data, etc.).
//
// Usage:
//
//	go run ./test_ct [flags]
//
// Flags:
//
//	-arch string
//	      Target architecture (GOARCH). Use -list-arch to see available architectures.
//	      (default: current GOARCH)
//	-os string
//	      Target operating system (GOOS). (default: "linux")
//	-func string
//	      Regex pattern to filter functions to analyze. (default: analyze all)
//	-module string
//	      Path to the Go module to analyze. (default: parent directory)
//	-warnings
//	      Include warning-level violations (e.g., conditional branches).
//	-json
//	      Output report in JSON format.
//	-github
//	      Output GitHub Actions annotations.
//	-verbose
//	      Enable verbose output.
//	-list-arch
//	      List supported architectures and exit.
//	-keep-binary
//	      Keep the compiled binary after analysis.
//
// Example:
//
//	# Analyze field operations on amd64
//	go run ./test_ct -arch=amd64 -func='field\.'
//
//	# Analyze all code on arm64 with warnings
//	go run ./test_ct -arch=arm64 -warnings
//
//	# JSON output for CI
//	go run ./test_ct -arch=amd64 -json
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/trailofbits/ml-dsa/test_ct/analyzer"
	"github.com/trailofbits/ml-dsa/test_ct/arch"
	"github.com/trailofbits/ml-dsa/test_ct/build"
	"github.com/trailofbits/ml-dsa/test_ct/disasm"

	// Import architecture implementations to register them
	_ "github.com/trailofbits/ml-dsa/test_ct/arch"
)

func main() {
	// Define flags
	targetArch := flag.String("arch", runtime.GOARCH, "Target architecture (GOARCH)")
	targetOS := flag.String("os", "linux", "Target operating system (GOOS)")
	funcPattern := flag.String("func", "", "Regex pattern to filter functions")
	modulePath := flag.String("module", "", "Path to Go module (default: parent directory)")
	includeWarnings := flag.Bool("warnings", false, "Include warning-level violations")
	jsonOutput := flag.Bool("json", false, "Output JSON format")
	githubOutput := flag.Bool("github", false, "Output GitHub Actions annotations")
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	listArch := flag.Bool("list-arch", false, "List supported architectures")
	keepBinary := flag.Bool("keep-binary", false, "Keep compiled binary after analysis")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "test_ct - Constant-Time Verification Tool\n\n")
		fmt.Fprintf(os.Stderr, "Analyzes Go code for instructions that could leak timing information.\n\n")
		fmt.Fprintf(os.Stderr, "Usage: go run ./test_ct [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  go run ./test_ct -arch=amd64 -func='field\\.'\n")
		fmt.Fprintf(os.Stderr, "  go run ./test_ct -arch=arm64 -warnings -json\n")
		fmt.Fprintf(os.Stderr, "  go run ./test_ct -list-arch\n")
	}

	flag.Parse()

	// Handle -list-arch
	if *listArch {
		printArchitectures()
		return
	}

	// Determine output format
	format := analyzer.FormatText
	if *jsonOutput {
		format = analyzer.FormatJSON
	} else if *githubOutput {
		format = analyzer.FormatGitHubActions
	}

	// Resolve module path
	modPath := *modulePath
	if modPath == "" {
		// Default to parent directory (assuming we're in test_ct/)
		exe, err := os.Executable()
		if err != nil {
			// Fall back to working directory
			wd, _ := os.Getwd()
			modPath = filepath.Dir(wd)
		} else {
			modPath = filepath.Dir(filepath.Dir(exe))
		}

		// If that doesn't work, try current directory's parent
		if _, err := os.Stat(filepath.Join(modPath, "go.mod")); err != nil {
			wd, _ := os.Getwd()
			// Check if we're in test_ct directory
			if filepath.Base(wd) == "test_ct" {
				modPath = filepath.Dir(wd)
			} else {
				modPath = wd
			}
		}
	}

	// Verify go.mod exists
	goModPath := filepath.Join(modPath, "go.mod")
	if _, err := os.Stat(goModPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: go.mod not found at %s\n", goModPath)
		fmt.Fprintf(os.Stderr, "Please specify -module flag or run from the module directory.\n")
		os.Exit(1)
	}

	// Get architecture
	architecture := arch.Get(*targetArch)
	if architecture == nil {
		fmt.Fprintf(os.Stderr, "Error: unsupported architecture: %s\n", *targetArch)
		fmt.Fprintf(os.Stderr, "Use -list-arch to see supported architectures.\n")
		os.Exit(1)
	}

	// Verify OS is supported for this architecture
	supported := false
	for _, os := range architecture.SupportedGOOS() {
		if os == *targetOS {
			supported = true
			break
		}
	}
	if !supported {
		fmt.Fprintf(os.Stderr, "Warning: %s may not be fully supported on %s/%s\n",
			architecture.Name(), *targetOS, *targetArch)
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "Module path: %s\n", modPath)
		fmt.Fprintf(os.Stderr, "Target: %s/%s\n", *targetOS, *targetArch)
		if *funcPattern != "" {
			fmt.Fprintf(os.Stderr, "Function filter: %s\n", *funcPattern)
		}
		fmt.Fprintf(os.Stderr, "\n")
	}

	// Step 1: Build
	if *verbose {
		fmt.Fprintf(os.Stderr, "Building for %s/%s...\n", *targetOS, *targetArch)
	}

	buildResult, err := build.Build(build.Config{
		ModulePath: modPath,
		GOOS:       *targetOS,
		GOARCH:     *targetArch,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Build failed: %v\n", err)
		os.Exit(1)
	}

	// Cleanup on exit unless -keep-binary
	if !*keepBinary {
		defer func() { _ = buildResult.Cleanup() }()
		defer func() { _ = build.CleanupSyntheticMain(modPath) }()
	} else if *verbose {
		fmt.Fprintf(os.Stderr, "Binary saved to: %s\n", buildResult.BinaryPath)
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "Build successful: %s\n\n", buildResult.BinaryPath)
	}

	// Step 2: Disassemble
	if *verbose {
		fmt.Fprintf(os.Stderr, "Disassembling...\n")
	}

	functions, err := disasm.Disassemble(disasm.Config{
		BinaryPath:      buildResult.BinaryPath,
		FunctionPattern: *funcPattern,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Disassembly failed: %v\n", err)
		os.Exit(1)
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "Found %d functions\n\n", len(functions))
	}

	// Step 3: Analyze
	if *verbose {
		fmt.Fprintf(os.Stderr, "Analyzing for constant-time violations...\n\n")
	}

	report := analyzer.Analyze(analyzer.Config{
		Architecture:    architecture,
		Functions:       functions,
		IncludeWarnings: *includeWarnings,
		Verbose:         *verbose,
	})

	report.GOOS = *targetOS

	// Step 4: Output report
	if err := analyzer.WriteReport(os.Stdout, report, format); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing report: %v\n", err)
		os.Exit(1)
	}

	// Exit with error if analysis failed
	if !report.Passed {
		os.Exit(1)
	}
}

func printArchitectures() {
	fmt.Println("Supported Architectures")
	fmt.Println("=======================")
	fmt.Println()

	for _, goarch := range arch.List() {
		a := arch.Get(goarch)
		fmt.Printf("%-10s  %s\n", goarch, a.Name())
		fmt.Printf("            Supported OS: %s\n", strings.Join(a.SupportedGOOS(), ", "))
		fmt.Println()
	}
}
