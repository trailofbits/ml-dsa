// Package arch defines the architecture interface for constant-time analysis
// and provides a registry for architecture-specific instruction analyzers.
package arch

import (
	"sort"
	"strings"
	"sync"
)

// Instruction represents a single disassembled instruction.
type Instruction struct {
	Address  uint64 // Memory address of the instruction
	Mnemonic string // Instruction mnemonic (e.g., "DIV", "IDIV")
	Operands string // Operand string (e.g., "AX, BX")
	Raw      string // Full raw instruction line from disassembly
}

// Violation represents a detected constant-time violation.
type Violation struct {
	InstructionIndex int    // Index in the instruction slice
	Instruction      string // The violating instruction
	Reason           string // Why this instruction is dangerous
	Severity         string // "error" or "warning"
}

// Pattern represents a multi-instruction pattern that may indicate a timing leak.
type Pattern struct {
	Name        string
	Description string
	Match       func(instructions []Instruction) []Violation
}

// Architecture defines the interface for architecture-specific analyzers.
type Architecture interface {
	// Name returns the human-readable name of the architecture.
	Name() string

	// GOARCH returns the GOARCH value for this architecture.
	GOARCH() string

	// SupportedGOOS returns the list of supported GOOS values for this architecture.
	SupportedGOOS() []string

	// IsDangerous checks if a single instruction is dangerous.
	// Returns (isDangerous, reason) where reason explains why the instruction is dangerous.
	IsDangerous(mnemonic string) (bool, string)

	// Patterns returns a list of multi-instruction patterns to check.
	Patterns() []Pattern

	// AnalyzeInstructions performs full analysis on a slice of instructions.
	// This includes both single-instruction checks and pattern matching.
	AnalyzeInstructions(instructions []Instruction) []Violation
}

// BaseArchitecture provides common functionality for architecture implementations.
type BaseArchitecture struct {
	name                 string
	goarch               string
	supportedOS          []string
	dangerousInstructions map[string]string // mnemonic -> reason
	warningInstructions  map[string]string // mnemonic -> reason (warnings, not errors)
	patterns             []Pattern
}

// NewBaseArchitecture creates a new BaseArchitecture.
func NewBaseArchitecture(name, goarch string, supportedOS []string) *BaseArchitecture {
	return &BaseArchitecture{
		name:                 name,
		goarch:               goarch,
		supportedOS:          supportedOS,
		dangerousInstructions: make(map[string]string),
		warningInstructions:  make(map[string]string),
		patterns:             nil,
	}
}

// Name returns the architecture name.
func (b *BaseArchitecture) Name() string {
	return b.name
}

// GOARCH returns the GOARCH value.
func (b *BaseArchitecture) GOARCH() string {
	return b.goarch
}

// SupportedGOOS returns supported OS list.
func (b *BaseArchitecture) SupportedGOOS() []string {
	return b.supportedOS
}

// AddDangerous adds a dangerous instruction to the list.
func (b *BaseArchitecture) AddDangerous(mnemonic, reason string) {
	b.dangerousInstructions[strings.ToUpper(mnemonic)] = reason
}

// AddWarning adds a warning-level instruction to the list.
func (b *BaseArchitecture) AddWarning(mnemonic, reason string) {
	b.warningInstructions[strings.ToUpper(mnemonic)] = reason
}

// AddPattern adds a multi-instruction pattern.
func (b *BaseArchitecture) AddPattern(p Pattern) {
	b.patterns = append(b.patterns, p)
}

// IsDangerous checks if an instruction mnemonic is dangerous.
func (b *BaseArchitecture) IsDangerous(mnemonic string) (bool, string) {
	upper := strings.ToUpper(mnemonic)
	if reason, ok := b.dangerousInstructions[upper]; ok {
		return true, reason
	}
	return false, ""
}

// IsWarning checks if an instruction mnemonic is a warning.
func (b *BaseArchitecture) IsWarning(mnemonic string) (bool, string) {
	upper := strings.ToUpper(mnemonic)
	if reason, ok := b.warningInstructions[upper]; ok {
		return true, reason
	}
	return false, ""
}

// Patterns returns the list of patterns.
func (b *BaseArchitecture) Patterns() []Pattern {
	return b.patterns
}

// AnalyzeInstructions performs full analysis.
func (b *BaseArchitecture) AnalyzeInstructions(instructions []Instruction) []Violation {
	var violations []Violation

	// Check individual instructions
	for i, inst := range instructions {
		if dangerous, reason := b.IsDangerous(inst.Mnemonic); dangerous {
			violations = append(violations, Violation{
				InstructionIndex: i,
				Instruction:      inst.Raw,
				Reason:           reason,
				Severity:         "error",
			})
		} else if warning, reason := b.IsWarning(inst.Mnemonic); warning {
			violations = append(violations, Violation{
				InstructionIndex: i,
				Instruction:      inst.Raw,
				Reason:           reason,
				Severity:         "warning",
			})
		}
	}

	// Check patterns
	for _, pattern := range b.patterns {
		patternViolations := pattern.Match(instructions)
		violations = append(violations, patternViolations...)
	}

	return violations
}

// Registry for architectures
var (
	registryMu sync.RWMutex
	registry   = make(map[string]Architecture)
)

// Register adds an architecture to the registry.
func Register(arch Architecture) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[arch.GOARCH()] = arch
}

// Get retrieves an architecture by GOARCH value.
func Get(goarch string) Architecture {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return registry[goarch]
}

// List returns all registered GOARCH values.
func List() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	var archs []string
	for k := range registry {
		archs = append(archs, k)
	}
	sort.Strings(archs)
	return archs
}

// ListArchitectures returns all registered architectures.
func ListArchitectures() []Architecture {
	registryMu.RLock()
	defer registryMu.RUnlock()
	var archs []Architecture
	for _, v := range registry {
		archs = append(archs, v)
	}
	return archs
}
