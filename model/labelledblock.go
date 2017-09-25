package model

// LabelledBlock groups OpaqueCode with its labels.
type LabelledBlock struct {
	labels []Label
	// prose is presumably human language documentation for the OpaqueCode.
	prose []byte
	code  OpaqueCode
}

func NewLabelledBlock(labels []Label, p []byte, c string) *LabelledBlock {
	return &LabelledBlock{labels, p, OpaqueCode(c)}
}
func (x *LabelledBlock) Labels() []Label  { return x.labels }
func (x *LabelledBlock) Prose() []byte    { return x.prose }
func (x *LabelledBlock) Code() OpaqueCode { return x.code }

func (x *LabelledBlock) HasLabel(label Label) bool {
	return hasLabel(x.Labels(), label)
}

func hasLabel(labels []Label, label Label) bool {
	for _, l := range labels {
		if l == label {
			return true
		}
	}
	return false
}
