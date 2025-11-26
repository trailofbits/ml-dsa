package arch

func init() {
	Register(newARM())
}

// arm implements the Architecture interface for ARM (32-bit).
type arm struct {
	*BaseArchitecture
}

func newARM() *arm {
	base := NewBaseArchitecture(
		"ARM (32-bit)",
		"arm",
		[]string{"linux", "freebsd", "openbsd", "netbsd", "android"},
	)

	// Division instructions - variable time
	base.AddDangerous("UDIV", "UDIV has data-dependent timing on ARM")
	base.AddDangerous("SDIV", "SDIV has data-dependent timing on ARM")

	// Multiplication - variable time on older ARM cores (ARM7, ARM9)
	// These are warnings because modern ARM cores (Cortex-A series) have constant-time MUL
	base.AddWarning("MUL", "MUL may have variable timing on older ARM cores (ARM7/ARM9)")
	base.AddWarning("MLA", "MLA may have variable timing on older ARM cores (ARM7/ARM9)")
	base.AddWarning("SMULL", "SMULL may have variable timing on older ARM cores")
	base.AddWarning("UMULL", "UMULL may have variable timing on older ARM cores")
	base.AddWarning("SMLAL", "SMLAL may have variable timing on older ARM cores")
	base.AddWarning("UMLAL", "UMLAL may have variable timing on older ARM cores")

	// Floating-point division and square root
	base.AddDangerous("VDIV.F32", "VDIV.F32 (VFP single-precision division) has variable latency")
	base.AddDangerous("VDIV.F64", "VDIV.F64 (VFP double-precision division) has variable latency")
	base.AddDangerous("VSQRT.F32", "VSQRT.F32 (VFP single-precision sqrt) has variable latency")
	base.AddDangerous("VSQRT.F64", "VSQRT.F64 (VFP double-precision sqrt) has variable latency")

	// Conditional branches
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

	return &arm{BaseArchitecture: base}
}
