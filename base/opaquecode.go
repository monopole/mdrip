package base

// OpaqueCode is an opaque, uninterpreted, unknown block of text that
// is presumably shell commands parsed from markdown.  Fed into a
// shell interpreter, the entire thing either succeeds or fails.
type OpaqueCode string

// String form of OpaqueCode.
func (c OpaqueCode) String() string { return string(c) }

// Bytes of the code.
func (c OpaqueCode) Bytes() []byte { return []byte(c) }

// NoCode is a constructor for NoCode - easy to search for usage.
func NoCode() OpaqueCode { return "" }
