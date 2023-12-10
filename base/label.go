package base

// Label is used to select code blocks, and group them into
// categories, e.g. run these blocks under test, run these blocks to do setup, etc.
type Label string

// String form of the label.
func (l Label) String() string { return string(l) }

const (
	// WildCardLabel matches any label.
	WildCardLabel = Label(`__wildcard__`)
	// AnonLabel may be used as a label placeholder when a label is needed but not specified.
	AnonLabel = Label(`__anonymous__`)
	// SleepLabel indicates the author wants a sleep after the block in a test context
	// where there is no natural human-caused pause.
	SleepLabel = Label(`sleep`)
)

// NoLabels is easier to read than the literal empty array.
func NoLabels() []Label { return []Label{} }
