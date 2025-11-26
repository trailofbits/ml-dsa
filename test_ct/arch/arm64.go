package arch

func init() {
	Register(newARM64())
}

// arm64 implements the Architecture interface for AArch64.
type arm64 struct {
	*BaseArchitecture
}

func newARM64() *arm64 {
	base := NewBaseArchitecture(
		"AArch64 (ARM64)",
		"arm64",
		[]string{"linux", "darwin", "windows", "freebsd", "openbsd", "netbsd", "ios", "android"},
	)

	// Division instructions - early termination optimization makes these variable-time
	// Note: Even with DIT (Data Independent Timing) enabled, division is NOT constant-time
	base.AddDangerous("UDIV", "UDIV has early termination optimization; execution time depends on operand values")
	base.AddDangerous("SDIV", "SDIV has early termination optimization; execution time depends on operand values")

	// Floating-point division - variable latency
	base.AddDangerous("FDIV", "FDIV (FP division) has variable latency based on operand values")
	base.AddDangerous("FDIVS", "FDIVS (single-precision FP division) has variable latency")
	base.AddDangerous("FDIVD", "FDIVD (double-precision FP division) has variable latency")

	// Square root - variable latency
	base.AddDangerous("FSQRT", "FSQRT has variable latency based on operand values")
	base.AddDangerous("FSQRTS", "FSQRTS (single-precision square root) has variable latency")
	base.AddDangerous("FSQRTD", "FSQRTD (double-precision square root) has variable latency")

	// Conditional branches - warnings only
	// Note: CSEL, CSINC, CSNEG, CSINV are safe constant-time conditional selects
	base.AddWarning("B.EQ", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("B.NE", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("B.CS", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("B.CC", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("B.MI", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("B.PL", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("B.VS", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("B.VC", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("B.HI", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("B.LS", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("B.GE", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("B.LT", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("B.GT", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("B.LE", "conditional branch may leak timing information if condition depends on secret data")

	// Compare and branch instructions
	base.AddWarning("CBZ", "compare-and-branch may leak timing information if value depends on secret data")
	base.AddWarning("CBNZ", "compare-and-branch may leak timing information if value depends on secret data")
	base.AddWarning("TBZ", "test-bit-and-branch may leak timing information if value depends on secret data")
	base.AddWarning("TBNZ", "test-bit-and-branch may leak timing information if value depends on secret data")

	// Go's assembler uses different mnemonics - add those too
	base.AddWarning("BEQ", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("BNE", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("BCS", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("BCC", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("BMI", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("BPL", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("BVS", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("BVC", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("BHI", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("BLS", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("BGE", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("BLT", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("BGT", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("BLE", "conditional branch may leak timing information if condition depends on secret data")

	return &arm64{BaseArchitecture: base}
}
