// Package build provides functionality to cross-compile Go modules for different
// target platforms for subsequent disassembly and analysis.
package build

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Config holds the configuration for building a Go module.
type Config struct {
	// ModulePath is the path to the Go module root directory.
	ModulePath string

	// OutputDir is the directory where the built binary will be placed.
	// If empty, a temporary directory will be used.
	OutputDir string

	// GOOS is the target operating system (e.g., "linux", "darwin", "windows").
	GOOS string

	// GOARCH is the target architecture (e.g., "amd64", "arm64", "386").
	GOARCH string

	// BuildFlags are additional flags to pass to go build.
	BuildFlags []string

	// Package is the package to build. If empty, builds "./..." from ModulePath.
	Package string
}

// Result contains the result of a build operation.
type Result struct {
	// BinaryPath is the path to the built binary.
	BinaryPath string

	// TempDir is the temporary directory created for the build (if any).
	// The caller is responsible for cleaning this up if non-empty.
	TempDir string
}

// Build cross-compiles a Go module for the specified target platform.
// It returns the path to the built binary and any error encountered.
//
// For library modules (those without a main package), this function builds
// a test binary using `go test -c` which includes all the code without
// dead code elimination.
func Build(cfg Config) (*Result, error) {
	if cfg.ModulePath == "" {
		return nil, fmt.Errorf("ModulePath is required")
	}

	if cfg.GOOS == "" {
		return nil, fmt.Errorf("GOOS is required")
	}

	if cfg.GOARCH == "" {
		return nil, fmt.Errorf("GOARCH is required")
	}

	// Resolve module path to absolute
	modulePath, err := filepath.Abs(cfg.ModulePath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve module path: %w", err)
	}

	// Create output directory if not specified
	var tempDir string
	outputDir := cfg.OutputDir
	if outputDir == "" {
		tempDir, err = os.MkdirTemp("", "ct-build-*")
		if err != nil {
			return nil, fmt.Errorf("failed to create temp directory: %w", err)
		}
		outputDir = tempDir
	}

	// Determine output binary name
	binaryName := fmt.Sprintf("ct-analyze-%s-%s", cfg.GOOS, cfg.GOARCH)
	if cfg.GOOS == "windows" {
		binaryName += ".exe"
	}
	binaryPath := filepath.Join(outputDir, binaryName)

	// Check if this is a library module (no main package)
	isLibrary := cfg.Package == "" && !hasMainPackage(modulePath)

	var cmd *exec.Cmd
	if isLibrary {
		// For library modules, use `go test -c` to build a test binary.
		// This includes all code without dead code elimination because tests
		// exercise the actual functions.
		// We build the test binary for a package that has comprehensive tests.
		testPkg := findBestTestPackage(modulePath)
		args := []string{"test", "-c", "-o", binaryPath}
		args = append(args, cfg.BuildFlags...)
		args = append(args, testPkg)
		cmd = exec.Command("go", args...)
	} else {
		// Determine what package to build
		pkg := cfg.Package
		if pkg == "" {
			pkg, err = findMainPackage(modulePath)
			if err != nil {
				if tempDir != "" {
					os.RemoveAll(tempDir)
				}
				return nil, err
			}
		}

		args := []string{"build"}
		args = append(args, cfg.BuildFlags...)
		args = append(args, "-o", binaryPath, pkg)
		cmd = exec.Command("go", args...)
	}

	cmd.Dir = modulePath
	cmd.Env = append(os.Environ(),
		"GOOS="+cfg.GOOS,
		"GOARCH="+cfg.GOARCH,
		"CGO_ENABLED=0", // Disable CGO for cross-compilation
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Clean up temp dir on error
		if tempDir != "" {
			os.RemoveAll(tempDir)
		}
		return nil, fmt.Errorf("go build failed: %w\nOutput: %s", err, string(output))
	}

	return &Result{
		BinaryPath: binaryPath,
		TempDir:    tempDir,
	}, nil
}

// hasMainPackage checks if the module has a main package.
func hasMainPackage(modulePath string) bool {
	// Check for cmd directory
	cmdDir := filepath.Join(modulePath, "cmd")
	if info, err := os.Stat(cmdDir); err == nil && info.IsDir() {
		entries, err := os.ReadDir(cmdDir)
		if err == nil && len(entries) > 0 {
			return true
		}
	}

	// Check for main.go in root
	mainFile := filepath.Join(modulePath, "main.go")
	if _, err := os.Stat(mainFile); err == nil {
		return true
	}

	return false
}

// findBestTestPackage finds the package with the most comprehensive tests.
// For crypto libraries, this is usually one of the main API packages that
// exercises all the internal code.
func findBestTestPackage(modulePath string) string {
	// Look for packages with test files, preferring top-level API packages
	// that are likely to exercise all internal code
	candidates := []string{
		"./mldsa87",   // ML-DSA specific - exercises all field/ring code
		"./mldsa65",
		"./mldsa44",
		"./internal/field",
		"./internal",
	}

	for _, candidate := range candidates {
		dir := filepath.Join(modulePath, candidate[2:]) // Remove "./" prefix
		if _, err := os.Stat(dir); err == nil {
			// Check if it has test files
			entries, err := os.ReadDir(dir)
			if err == nil {
				for _, entry := range entries {
					if strings.HasSuffix(entry.Name(), "_test.go") {
						return candidate
					}
				}
			}
		}
	}

	// Fallback: look for any package with test files
	var testPkg string
	_ = filepath.Walk(modulePath, func(path string, info os.FileInfo, err error) error {
		if err != nil || testPkg != "" {
			return err
		}
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "vendor" || name == "testdata" || name == "test_ct" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, "_test.go") {
			dir := filepath.Dir(path)
			relDir, err := filepath.Rel(modulePath, dir)
			if err == nil {
				testPkg = "./" + filepath.ToSlash(relDir)
			}
			return filepath.SkipAll
		}
		return nil
	})

	if testPkg != "" {
		return testPkg
	}

	// Last resort: just use the root package
	return "."
}

// findMainPackage looks for a main package in the module that can be built
// into an executable. For library modules without a cmd directory, it creates
// a synthetic main package that references all exported functions to prevent
// dead code elimination.
func findMainPackage(modulePath string) (string, error) {
	// Check common locations for main packages
	candidates := []string{
		"./cmd/...",
		"./main.go",
		".",
	}

	for _, candidate := range candidates {
		if candidate == "./cmd/..." {
			cmdDir := filepath.Join(modulePath, "cmd")
			if info, err := os.Stat(cmdDir); err == nil && info.IsDir() {
				entries, err := os.ReadDir(cmdDir)
				if err == nil && len(entries) > 0 {
					// Found cmd directory with subdirectories
					for _, entry := range entries {
						if entry.IsDir() {
							return "./cmd/" + entry.Name(), nil
						}
					}
				}
			}
		} else if candidate == "./main.go" {
			mainFile := filepath.Join(modulePath, "main.go")
			if _, err := os.Stat(mainFile); err == nil {
				return ".", nil
			}
		}
	}

	// No main package found - create a synthetic one that imports the library
	// This allows us to build the library code into an executable for disassembly
	return createSyntheticMain(modulePath)
}

// createSyntheticMain creates a synthetic main package that imports all
// packages in the module, allowing us to build and disassemble library code.
func createSyntheticMain(modulePath string) (string, error) {
	// Read go.mod to get module name
	goModPath := filepath.Join(modulePath, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return "", fmt.Errorf("failed to read go.mod: %w", err)
	}

	// Parse module name from go.mod
	moduleName := ""
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			moduleName = strings.TrimPrefix(line, "module ")
			break
		}
	}
	if moduleName == "" {
		return "", fmt.Errorf("could not find module name in go.mod")
	}

	// Find all Go packages in the module
	packages, err := findGoPackages(modulePath, moduleName)
	if err != nil {
		return "", fmt.Errorf("failed to find packages: %w", err)
	}

	if len(packages) == 0 {
		return "", fmt.Errorf("no Go packages found in module")
	}

	// Create synthetic main directory
	synthDir := filepath.Join(modulePath, ".ct-synthetic-main")
	if err := os.MkdirAll(synthDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create synthetic main directory: %w", err)
	}

	// Generate main.go that imports all packages
	var imports strings.Builder
	imports.WriteString("package main\n\n")
	imports.WriteString("import (\n")
	for _, pkg := range packages {
		imports.WriteString(fmt.Sprintf("\t_ %q\n", pkg))
	}
	imports.WriteString(")\n\n")
	imports.WriteString("func main() {}\n")

	mainPath := filepath.Join(synthDir, "main.go")
	if err := os.WriteFile(mainPath, []byte(imports.String()), 0644); err != nil {
		return "", fmt.Errorf("failed to write synthetic main.go: %w", err)
	}

	return "./.ct-synthetic-main", nil
}

// findGoPackages finds all Go packages in the module.
func findGoPackages(modulePath, moduleName string) ([]string, error) {
	var packages []string

	err := filepath.Walk(modulePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and common non-code directories
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "vendor" || name == "testdata" {
				return filepath.SkipDir
			}
			// Skip the test_ct directory itself - it's a main package and can't be imported
			if name == "test_ct" {
				return filepath.SkipDir
			}
			return nil
		}

		// Check for .go files (not test files)
		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			dir := filepath.Dir(path)
			relDir, err := filepath.Rel(modulePath, dir)
			if err != nil {
				return nil
			}

			// Convert to import path
			var importPath string
			if relDir == "." {
				importPath = moduleName
			} else {
				importPath = moduleName + "/" + filepath.ToSlash(relDir)
			}

			// Skip packages in test_ct (they're main packages or part of the tool)
			if strings.Contains(importPath, "/test_ct") {
				return nil
			}

			// Check if we already added this package
			found := false
			for _, p := range packages {
				if p == importPath {
					found = true
					break
				}
			}
			if !found {
				packages = append(packages, importPath)
			}
		}
		return nil
	})

	return packages, err
}

// Cleanup removes any temporary files created during the build.
func (r *Result) Cleanup() error {
	if r.TempDir != "" {
		return os.RemoveAll(r.TempDir)
	}
	return nil
}

// CleanupSyntheticMain removes the synthetic main package if it was created.
func CleanupSyntheticMain(modulePath string) error {
	synthDir := filepath.Join(modulePath, ".ct-synthetic-main")
	if _, err := os.Stat(synthDir); err == nil {
		return os.RemoveAll(synthDir)
	}
	return nil
}
