package arch

func init() {
	Register(newAMD64())
}

// amd64 implements the Architecture interface for x86-64.
type amd64 struct {
	*BaseArchitecture
}

func newAMD64() *amd64 {
	base := NewBaseArchitecture(
		"x86-64 (AMD64)",
		"amd64",
		[]string{"linux", "darwin", "windows", "freebsd", "openbsd", "netbsd"},
	)

	// Division instructions - always variable time based on operand values
	base.AddDangerous("DIV", "DIV has data-dependent timing; execution time varies based on operand values")
	base.AddDangerous("IDIV", "IDIV has data-dependent timing; execution time varies based on operand values")
	base.AddDangerous("DIVB", "DIVB has data-dependent timing; execution time varies based on operand values")
	base.AddDangerous("DIVW", "DIVW has data-dependent timing; execution time varies based on operand values")
	base.AddDangerous("DIVL", "DIVL has data-dependent timing; execution time varies based on operand values")
	base.AddDangerous("DIVQ", "DIVQ has data-dependent timing; execution time varies based on operand values")
	base.AddDangerous("IDIVB", "IDIVB has data-dependent timing; execution time varies based on operand values")
	base.AddDangerous("IDIVW", "IDIVW has data-dependent timing; execution time varies based on operand values")
	base.AddDangerous("IDIVL", "IDIVL has data-dependent timing; execution time varies based on operand values")
	base.AddDangerous("IDIVQ", "IDIVQ has data-dependent timing; execution time varies based on operand values")

	// Floating-point division - variable latency
	base.AddDangerous("DIVSS", "DIVSS (scalar single FP division) has variable latency")
	base.AddDangerous("DIVSD", "DIVSD (scalar double FP division) has variable latency")
	base.AddDangerous("DIVPS", "DIVPS (packed single FP division) has variable latency")
	base.AddDangerous("DIVPD", "DIVPD (packed double FP division) has variable latency")
	base.AddDangerous("VDIVSS", "VDIVSS (AVX scalar single FP division) has variable latency")
	base.AddDangerous("VDIVSD", "VDIVSD (AVX scalar double FP division) has variable latency")
	base.AddDangerous("VDIVPS", "VDIVPS (AVX packed single FP division) has variable latency")
	base.AddDangerous("VDIVPD", "VDIVPD (AVX packed double FP division) has variable latency")

	// Square root - variable latency
	base.AddDangerous("SQRTSS", "SQRTSS has variable latency based on operand values")
	base.AddDangerous("SQRTSD", "SQRTSD has variable latency based on operand values")
	base.AddDangerous("SQRTPS", "SQRTPS has variable latency based on operand values")
	base.AddDangerous("SQRTPD", "SQRTPD has variable latency based on operand values")
	base.AddDangerous("VSQRTSS", "VSQRTSS has variable latency based on operand values")
	base.AddDangerous("VSQRTSD", "VSQRTSD has variable latency based on operand values")
	base.AddDangerous("VSQRTPS", "VSQRTPS has variable latency based on operand values")
	base.AddDangerous("VSQRTPD", "VSQRTPD has variable latency based on operand values")

	// Conditional branches - warnings only (may be on public data)
	// These are tracked as warnings because they could be branching on public data
	base.AddWarning("JE", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("JNE", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("JZ", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("JNZ", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("JA", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("JAE", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("JB", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("JBE", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("JG", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("JGE", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("JL", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("JLE", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("JO", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("JNO", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("JS", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("JNS", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("JP", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("JNP", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("JC", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("JNC", "conditional branch may leak timing information if condition depends on secret data")

	return &amd64{BaseArchitecture: base}
}
