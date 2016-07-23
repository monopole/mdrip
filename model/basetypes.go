package model

type FileName string

// Labels are applied to code blocks to identify them and allow the
// blocks to be grouped into categories, e.g. tests or tutorials.
type Label string

const (
	AnyLabel = Label(`__AnyLabel__`)
)

func (l Label) String() string {
	return string(l)
}

func (l Label) IsAny() bool {
	return l == AnyLabel
}
