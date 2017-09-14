package model

import (
	"io/ioutil"
	"path/filepath"

	"os"
)

type FileName string

func (n FileName) ReadDir() ([]os.FileInfo, error) {
	return ioutil.ReadDir(string(n))
}

func (n FileName) Base() string {
	return filepath.Base(string(n))
}

func (n FileName) Join(info os.FileInfo) FileName {
	return FileName(filepath.Join(string(n), info.Name()))
}

func (n FileName) Read() (string, error) {
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
	AnyLabel = Label(`__AnyLabel__`)
)

func (l Label) String() string {
	return string(l)
}

func (l Label) IsAny() bool {
	return l == AnyLabel
}
