package arch

func init() {
	Register(newS390X())
}

// s390x implements the Architecture interface for IBM z/Architecture (s390x).
type s390x struct {
	*BaseArchitecture
}

func newS390X() *s390x {
	base := NewBaseArchitecture(
		"IBM z/Architecture (s390x)",
		"s390x",
		[]string{"linux"},
	)

	// Division instructions - variable time
	// s390x division instructions have data-dependent timing
	base.AddDangerous("D", "D (divide) has data-dependent timing")
	base.AddDangerous("DR", "DR (divide register) has data-dependent timing")
	base.AddDangerous("DL", "DL (divide logical) has data-dependent timing")
	base.AddDangerous("DLR", "DLR (divide logical register) has data-dependent timing")
	base.AddDangerous("DLG", "DLG (divide logical 64-bit) has data-dependent timing")
	base.AddDangerous("DLGR", "DLGR (divide logical register 64-bit) has data-dependent timing")
	base.AddDangerous("DSG", "DSG (divide single 64-bit) has data-dependent timing")
	base.AddDangerous("DSGR", "DSGR (divide single register 64-bit) has data-dependent timing")
	base.AddDangerous("DSGF", "DSGF (divide single 64/32-bit) has data-dependent timing")
	base.AddDangerous("DSGFR", "DSGFR (divide single register 64/32-bit) has data-dependent timing")

	// Floating-point division
	base.AddDangerous("DDB", "DDB (divide FP long) has variable latency")
	base.AddDangerous("DDBR", "DDBR (divide FP long register) has variable latency")
	base.AddDangerous("DEB", "DEB (divide FP short) has variable latency")
	base.AddDangerous("DEBR", "DEBR (divide FP short register) has variable latency")
	base.AddDangerous("DXB", "DXB (divide FP extended) has variable latency")
	base.AddDangerous("DXBR", "DXBR (divide FP extended register) has variable latency")

	// Floating-point square root
	base.AddDangerous("SQDB", "SQDB (square root FP long) has variable latency")
	base.AddDangerous("SQDBR", "SQDBR (square root FP long register) has variable latency")
	base.AddDangerous("SQEB", "SQEB (square root FP short) has variable latency")
	base.AddDangerous("SQEBR", "SQEBR (square root FP short register) has variable latency")
	base.AddDangerous("SQXB", "SQXB (square root FP extended) has variable latency")
	base.AddDangerous("SQXBR", "SQXBR (square root FP extended register) has variable latency")

	// Conditional branches
	base.AddWarning("BC", "conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("BCR", "conditional branch register may leak timing information if condition depends on secret data")
	base.AddWarning("BRC", "relative conditional branch may leak timing information if condition depends on secret data")
	base.AddWarning("BRCL", "relative conditional branch long may leak timing information if condition depends on secret data")
	base.AddWarning("BRE", "branch on equal may leak timing information if condition depends on secret data")
	base.AddWarning("BRNE", "branch on not equal may leak timing information if condition depends on secret data")
	base.AddWarning("BRH", "branch on high may leak timing information if condition depends on secret data")
	base.AddWarning("BRNH", "branch on not high may leak timing information if condition depends on secret data")
	base.AddWarning("BRL", "branch on low may leak timing information if condition depends on secret data")
	base.AddWarning("BRNL", "branch on not low may leak timing information if condition depends on secret data")

	// Compare and branch
	base.AddWarning("CRJ", "compare and branch relative may leak timing information")
	base.AddWarning("CGRJ", "compare 64-bit and branch relative may leak timing information")
	base.AddWarning("CIJ", "compare immediate and branch relative may leak timing information")
	base.AddWarning("CGIJ", "compare 64-bit immediate and branch relative may leak timing information")
	base.AddWarning("CLRJ", "compare logical and branch relative may leak timing information")
	base.AddWarning("CLGRJ", "compare logical 64-bit and branch relative may leak timing information")

	return &s390x{BaseArchitecture: base}
}
