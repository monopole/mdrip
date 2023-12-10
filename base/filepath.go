package base

import (
	"os"
	"path/filepath"
	"strings"
)

// FilePath holds a non-empty, absolute file path.
type FilePath string

// ReadDir treats the path as a directory name and reads it.
func (n FilePath) ReadDir() ([]os.DirEntry, error) {
	return os.ReadDir(string(n))
}

// IsEmpty is true if the path is empty
func (n FilePath) IsEmpty() bool {
	return n == ""
}

// Base returns the last part of the file path, e.g. foo in /usr/local/foo.
func (n FilePath) Base() string {
	arg := string(n)
	ext := filepath.Ext(arg)
	if len(ext) < 1 || ext != ".md" {
		return filepath.Base(arg)
	}
	return filepath.Base(arg[:strings.Index(arg, ext)])

}

// Join joins two file name parts using the OS-specific file path joiner.
func (n FilePath) Join(e os.DirEntry) FilePath {
	return FilePath(filepath.Join(string(n), e.Name()))
}

// Read reads the file into a string.
func (n FilePath) Read() (string, error) {
	contents, err := os.ReadFile(string(n))
	if err != nil {
		return "", err
	}
	return string(contents), nil
}

// IsDesirableFile returns true if the filepath seems like markdown.
func (n FilePath) IsDesirableFile() bool {
	s, err := os.Stat(string(n))
	if err != nil {
		return false
	}
	if s.IsDir() {
		return false
	}
	if !s.Mode().IsRegular() {
		return false
	}
	if filepath.Ext(s.Name()) != ".md" {
		return false
	}
	base := filepath.Base(s.Name())
	const badLeadingChar = "~.#"
	if strings.Index(badLeadingChar, string(base[0])) > -1 {
		return false
	}
	return true
}

// IsDesirableDir returns true if the directory should be processed.
func (n FilePath) IsDesirableDir() bool {
	s, err := os.Stat(string(n))
	if err != nil {
		return false
	}
	if !s.IsDir() {
		return false
	}
	// Allow special dir names.
	if s.Name() == "." || s.Name() == "./" || s.Name() == ".." {
		return true
	}
	// Ignore .git, etc.
	if strings.HasPrefix(filepath.Base(s.Name()), ".") {
		return false
	}
	return true
}

// IsOrderFile returns true if the file appears to be a "reorder"
// file specifying how to re-order the files in the directory
// in some fashion other than directory order.
func (n FilePath) IsOrderFile() bool {
	s, err := os.Stat(string(n))
	if err != nil {
		return false
	}
	if s.IsDir() {
		return false
	}
	if !s.Mode().IsRegular() {
		return false
	}
	return filepath.Base(s.Name()) == "README_ORDER.txt"
}
