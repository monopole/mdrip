package model

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

type TypeSessId string

type FilePath string

func (n FilePath) ReadDir() ([]os.FileInfo, error) {
	return ioutil.ReadDir(string(n))
}

func (n FilePath) Base() string {
	return filepath.Base(string(n))
}

func (n FilePath) Join(info os.FileInfo) FilePath {
	return FilePath(filepath.Join(string(n), info.Name()))
}

func (n FilePath) Read() (string, error) {
	contents, err := ioutil.ReadFile(string(n))
	if err != nil {
		return "", err
	}
	return string(contents), nil
}

// Labels are applied to code blocks to identify them and allow the
// blocks to be grouped into categories, e.g. tests or tutorials.
type Label string

const (
	AnyLabel     = Label(`__AnyLabel__`)
	MistakeLabel = Label(`__MistakeLabel__`)
	SleepLabel   = Label(`sleep`)
)

func (l Label) String() string { return string(l) }
func (l Label) IsAny() bool    { return l == AnyLabel }

// OpaqueCode is an opaque, uninterpreted, unknown block of text that
// is presumably shell commands parsed from markdown.  Fed into a
// shell interpreter, the entire thing either succeeds or fails.
type OpaqueCode string

func (c OpaqueCode) String() string { return string(c) }
func (c OpaqueCode) Bytes() []byte  { return []byte(c) }

// Block groups OpaqueCode with its labels.
type Block struct {
	labels []Label
	// prose is presumably human language documentation for the OpaqueCode.
	prose []byte
	code  OpaqueCode
}

func shouldSleep(labels []Label) bool {
	for _, l := range labels {
		if l == SleepLabel {
			return true
		}
	}
	return false
}

func NewBlock(labels []Label, p string, c string) *Block {
	// If the command block has a 'sleep' label, add a brief sleep
	// at the end.  This hack give servers placed in the
	// background time to start, assuming they can do so in 2s!  Yeah, bad.
	if shouldSleep(labels) {
		c += "sleep 2s # Added by mdrip\n"
	}
	// Always add AnyLabel as a cheap matcher.
	labels = append(labels, AnyLabel)
	return &Block{labels, []byte(p), OpaqueCode(c)}
}
func (x *Block) Labels() []Label  { return x.labels }
func (x *Block) Prose() []byte    { return x.prose }
func (x *Block) Code() OpaqueCode { return x.code }
