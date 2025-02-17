package loader

// Label is used to select code blocks, and group them into
// categories, e.g. run these blocks under test, run these blocks to do setup, etc.
type Label string

// String form of the label.
func (l Label) String() string { return string(l) }

const (
	// WildCardLabel matches any label.
	WildCardLabel = Label(`__wildcard__`)

	// AnonLabel may be used as a label placeholder when a label is
	// needed but not specified.
	AnonLabel = Label(`__anonymous__`)

	// SleepLabel indicates the author wants a sleep after the block in a
	// test context where there is no natural human-caused pause.
	SleepLabel = Label(`sleep`)

	// SkipLabel is used on blocks that should be skipped in some context.
	SkipLabel = Label(`skip`)
)

type LabelList []Label

func NewLabelList(cbs []*CodeBlock) LabelList {
	labels := make([]Label, len(cbs))
	for j, block := range cbs {
		labels[j] = block.FirstLabel()
	}
	return labels
}

func (lst LabelList) Contains(l Label) bool {
	for i := range lst {
		if lst[i] == l {
			return true
		}
	}
	return false
}

// Equals is true if the slices have the same contents, ordering irrelevant.
func (lst LabelList) Equals(other LabelList) bool {
	if len(lst) != len(other) {
		return false
	}
	for i := range other {
		if !lst.Contains(other[i]) {
			return false
		}
	}
	return true
}
