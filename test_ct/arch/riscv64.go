package arch

func init() {
	Register(newRISCV64())
}

// riscv64 implements the Architecture interface for RISC-V 64-bit.
type riscv64 struct {
	*BaseArchitecture
}

func newRISCV64() *riscv64 {
	base := NewBaseArchitecture(
		"RISC-V 64-bit",
		"riscv64",
		[]string{"linux"},
	)

	// Integer division and remainder (M extension) - variable time
	// RISC-V does not guarantee constant-time division
	base.AddDangerous("DIV", "DIV has data-dependent timing on RISC-V")
	base.AddDangerous("DIVU", "DIVU has data-dependent timing on RISC-V")
	base.AddDangerous("REM", "REM has data-dependent timing on RISC-V")
	base.AddDangerous("REMU", "REMU has data-dependent timing on RISC-V")
	base.AddDangerous("DIVW", "DIVW (32-bit signed division) has data-dependent timing")
	base.AddDangerous("DIVUW", "DIVUW (32-bit unsigned division) has data-dependent timing")
	base.AddDangerous("REMW", "REMW (32-bit signed remainder) has data-dependent timing")
	base.AddDangerous("REMUW", "REMUW (32-bit unsigned remainder) has data-dependent timing")

	// Go assembler uses uppercase for RISC-V mnemonics
	base.AddDangerous("div", "div has data-dependent timing on RISC-V")
	base.AddDangerous("divu", "divu has data-dependent timing on RISC-V")
	base.AddDangerous("rem", "rem has data-dependent timing on RISC-V")
	base.AddDangerous("remu", "remu has data-dependent timing on RISC-V")
	base.AddDangerous("divw", "divw (32-bit signed division) has data-dependent timing")
	base.AddDangerous("divuw", "divuw (32-bit unsigned division) has data-dependent timing")
	base.AddDangerous("remw", "remw (32-bit signed remainder) has data-dependent timing")
	base.AddDangerous("remuw", "remuw (32-bit unsigned remainder) has data-dependent timing")

	// Floating-point division (F/D extensions)
	base.AddDangerous("FDIV.S", "FDIV.S (single-precision FP division) has variable latency")
	base.AddDangerous("FDIV.D", "FDIV.D (double-precision FP division) has variable latency")
	base.AddDangerous("FDIV.Q", "FDIV.Q (quad-precision FP division) has variable latency")
	base.AddDangerous("fdiv.s", "fdiv.s (single-precision FP division) has variable latency")
	base.AddDangerous("fdiv.d", "fdiv.d (double-precision FP division) has variable latency")
	base.AddDangerous("fdiv.q", "fdiv.q (quad-precision FP division) has variable latency")

	// Floating-point square root
	base.AddDangerous("FSQRT.S", "FSQRT.S (single-precision sqrt) has variable latency")
	base.AddDangerous("FSQRT.D", "FSQRT.D (double-precision sqrt) has variable latency")
	base.AddDangerous("FSQRT.Q", "FSQRT.Q (quad-precision sqrt) has variable latency")
	base.AddDangerous("fsqrt.s", "fsqrt.s (single-precision sqrt) has variable latency")
	base.AddDangerous("fsqrt.d", "fsqrt.d (double-precision sqrt) has variable latency")
	base.AddDangerous("fsqrt.q", "fsqrt.q (quad-precision sqrt) has variable latency")

	// Conditional branches
	base.AddWarning("BEQ", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("BNE", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("BLT", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("BGE", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("BLTU", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("BGEU", "conditional branch may leak timing information if condition depends on secret data")

	// Lowercase variants
	base.AddWarning("beq", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("bne", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("blt", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("bge", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("bltu", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("bgeu", "conditional branch may leak timing information if condition depends on secret data")

	return &riscv64{BaseArchitecture: base}
}
