package arch

func init() {
	Register(newPPC64())
	Register(newPPC64LE())
}

// ppc64 implements the Architecture interface for PowerPC 64-bit (big-endian).
type ppc64 struct {
	*BaseArchitecture
}

func newPPC64() *ppc64 {
	base := NewBaseArchitecture(
		"PowerPC 64-bit (Big Endian)",
		"ppc64",
		[]string{"linux", "aix"},
	)
	addPPCDangerousInstructions(base)
	return &ppc64{BaseArchitecture: base}
}

// ppc64le implements the Architecture interface for PowerPC 64-bit (little-endian).
type ppc64le struct {
	*BaseArchitecture
}

func newPPC64LE() *ppc64le {
	base := NewBaseArchitecture(
		"PowerPC 64-bit (Little Endian)",
		"ppc64le",
		[]string{"linux"},
	)
	addPPCDangerousInstructions(base)
	return &ppc64le{BaseArchitecture: base}
}

// addPPCDangerousInstructions adds common PowerPC dangerous instructions.
func addPPCDangerousInstructions(base *BaseArchitecture) {
	// Integer division - variable time
	base.AddDangerous("DIVW", "DIVW (signed 32-bit division) has data-dependent timing")
	base.AddDangerous("DIVWU", "DIVWU (unsigned 32-bit division) has data-dependent timing")
	base.AddDangerous("DIVD", "DIVD (signed 64-bit division) has data-dependent timing")
	base.AddDangerous("DIVDU", "DIVDU (unsigned 64-bit division) has data-dependent timing")
	base.AddDangerous("DIVWE", "DIVWE (signed word extended division) has data-dependent timing")
	base.AddDangerous("DIVWEU", "DIVWEU (unsigned word extended division) has data-dependent timing")
	base.AddDangerous("DIVDE", "DIVDE (signed doubleword extended division) has data-dependent timing")
	base.AddDangerous("DIVDEU", "DIVDEU (unsigned doubleword extended division) has data-dependent timing")

	// Go assembler uses lowercase for PowerPC
	base.AddDangerous("divw", "divw (signed 32-bit division) has data-dependent timing")
	base.AddDangerous("divwu", "divwu (unsigned 32-bit division) has data-dependent timing")
	base.AddDangerous("divd", "divd (signed 64-bit division) has data-dependent timing")
	base.AddDangerous("divdu", "divdu (unsigned 64-bit division) has data-dependent timing")

	// Modulo instructions (Power ISA 3.0+)
	base.AddDangerous("MODSW", "MODSW has data-dependent timing")
	base.AddDangerous("MODUW", "MODUW has data-dependent timing")
	base.AddDangerous("MODSD", "MODSD has data-dependent timing")
	base.AddDangerous("MODUD", "MODUD has data-dependent timing")

	// Floating-point division
	base.AddDangerous("FDIV", "FDIV (FP division) has variable latency")
	base.AddDangerous("FDIVS", "FDIVS (single-precision FP division) has variable latency")
	base.AddDangerous("fdiv", "fdiv (FP division) has variable latency")
	base.AddDangerous("fdivs", "fdivs (single-precision FP division) has variable latency")

	// Floating-point square root
	base.AddDangerous("FSQRT", "FSQRT has variable latency")
	base.AddDangerous("FSQRTS", "FSQRTS (single-precision sqrt) has variable latency")
	base.AddDangerous("fsqrt", "fsqrt has variable latency")
	base.AddDangerous("fsqrts", "fsqrts (single-precision sqrt) has variable latency")

	// Conditional branches
	base.AddWarning("BC", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("BCA", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("BCL", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("BCLA", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("BEQ", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("BNE", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("BLT", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("BGT", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("BLE", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("BGE", "conditional branch may leak timing information if condition depends on secret data")

	// Lowercase variants
	base.AddWarning("bc", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("beq", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("bne", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("blt", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("bgt", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("ble", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("bge", "conditional branch may leak timing information if condition depends on secret data")
}
