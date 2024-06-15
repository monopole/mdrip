package loader

import (
	"os"
	"strings"

	"github.com/spf13/afero"
)

// IsOrderingFile returns true if the file appears to be an "ordering" file
// specifying which files should come first in a directory.
func IsOrderingFile(info os.FileInfo) bool {
	if info.IsDir() {
		return false
	}
	if !info.Mode().IsRegular() {
		return false
	}
	return info.Name() == OrderingFileName
}

// LoadOrderFile returns a list of names specify file name order priority.
func LoadOrderFile(fs *afero.Afero, path string) ([]string, error) {
	contents, err := fs.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(contents), "\n"), nil
}

func ReorderFolders(x []*MyFolder, ordering []string) []*MyFolder {
	for i := len(ordering) - 1; i >= 0; i-- {
		x = shiftFolderToTop(x, ordering[i])
	}
	return x
}

func shiftFolderToTop(x []*MyFolder, top string) []*MyFolder {
	var first []*MyFolder
	var remainder []*MyFolder
	for _, f := range x {
		if f.Name() == top {
			first = append(first, f)
		} else {
			remainder = append(remainder, f)
		}
	}
	return append(first, remainder...)
}

func ReorderFiles(x []*MyFile, ordering []string) []*MyFile {
	for i := len(ordering) - 1; i >= 0; i-- {
		x = shiftFileToTop(x, ordering[i])
	}
	return shiftFileToTop(x, "README.md")
}

func shiftFileToTop(x []*MyFile, top string) []*MyFile {
	var first []*MyFile
	var remainder []*MyFile
	for _, f := range x {
		if f.Name() == top {
			first = append(first, f)
		} else {
			remainder = append(remainder, f)
		}
	}
	return append(first, remainder...)
}
