package base

// MdProse is documentation (plain text or markdown) for OpaqueCode.
type MdProse []byte

// String form of MdProse.
func (x MdProse) String() string { return string(x) }

// Bytes of MdProse.
func (x MdProse) Bytes() []byte { return []byte(x) }

// NoProse is placeholder for no prose.
func NoProse() MdProse { return []byte{} }
