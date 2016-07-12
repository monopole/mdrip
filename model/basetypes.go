package model

type FileName string

// Labels are applied to code blocks to identify them and allow the
// blocks to be grouped into categories, e.g. tests or tutorials.
type Label string

func (l Label) String() string {
	return string(l)
}
