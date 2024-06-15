package loader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// filter returns an error if conditions not met
type filter func(info os.FileInfo) error

var NotMarkDownErr = fmt.Errorf("not a simple markdown file")

// IsMarkDownFile passes markdown files.
func IsMarkDownFile(info os.FileInfo) error {
	if !info.Mode().IsRegular() {
		return NotMarkDownErr
	}
	if filepath.Ext(info.Name()) != ".md" {
		return NotMarkDownErr
	}
	const badLeadingChar = "~.#"
	if strings.Index(badLeadingChar, string(info.Name()[0])) >= 0 {
		return NotMarkDownErr
	}
	return nil
}

var IsADotDirErr = fmt.Errorf("not allowed to load from dot folder")

// IsNotADotDir returns an error if FileInfo happens
// to be a dot directory (.git, .config, etc.)
func IsNotADotDir(info os.FileInfo) error {
	n := FilePath(info.Name())
	// Allow special dir names.
	if n == CurrentDir || n == selfPath || n == upDir {
		return nil
	}
	// Ignore .git, etc.
	if len(n) > 1 && n[0] == '.' {
		return IsADotDirErr
	}
	return nil
}
