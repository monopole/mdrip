package loader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FsFilter returns an error if conditions for a file or folder are not met
type FsFilter func(info os.FileInfo) error

var NotMarkDownErr = fmt.Errorf("not a simple markdown file")

// IsWhatever passes everything.
func IsWhatever(_ os.FileInfo) error {
	return nil
}

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

var IsANodeCache = fmt.Errorf("not allowed to load from node cache")

// InNotIgnorableFolder returns an error if FileInfo happens
// to be a dot directory (.git, .config, etc.)
// TODO: write something that honors a local .gitignore
func InNotIgnorableFolder(info os.FileInfo) error {
	n := FilePath(info.Name())
	// Allow special dir names.
	if n == CurrentDir || n == selfPath || n == upDir {
		return nil
	}
	// Ignore .git, etc.
	if len(n) > 1 && n[0] == '.' {
		return IsADotDirErr
	}
	if n == "node_modules" {
		return IsANodeCache
	}
	return nil
}
