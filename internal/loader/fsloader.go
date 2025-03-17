package loader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"log/slog"
	"github.com/spf13/afero"
)

// FsLoader navigates and reads a file system.
type FsLoader struct {
	IsAllowedFile, IsAllowedFolder FsFilter
	fs                             *afero.Afero
}

// New returns a file system (FS) loader with default filters.
// For an in-memory FS, inject afero.NewMemMapFs().
// For a "real" disk-based system, inject afero.NewOsFs().
func New(fs afero.Fs, allowedFile FsFilter, allowedFolder FsFilter) *FsLoader {
	return &FsLoader{
		IsAllowedFile:   allowedFile,
		IsAllowedFolder: allowedFolder,
		fs:              &afero.Afero{Fs: fs},
	}
}

const (
	ReadmeFileName   = "README.md"
	OrderingFileName = "README_ORDER.txt"
	RootSlash        = FilePath(filepath.Separator)
	CurrentDir       = FilePath(".")
	selfPath         = CurrentDir + RootSlash
	upDir            = FilePath("..")
)

// LoadTrees loads several paths, wrapping them all in virtual folder.
func (fsl *FsLoader) LoadTrees(args []string) (*MyFolder, error) {
	if len(args) < 2 {
		slog.Warn("Warning, processing ALL directories because no input provided")
		arg := CurrentDir // By default, read the current directory.
		if len(args) == 1 {
			arg = FilePath(args[0])
		}
		return fsl.LoadOneTree(arg)
	}
	// Make one folder to hold all the argument folders.
	wrapper := NewFolder("virtual")
	for i := range args {
		fld, err := fsl.LoadOneTree(FilePath(args[i]))
		if err != nil {
			return nil, err
		}
		if fld != nil {
			wrapper.AddFolder(fld)
		}
	}
	if wrapper.IsEmpty() {
		return nil, nil
	}
	return wrapper, nil
}

// LoadOneTree loads a file tree from disk, possibly after first cloning a repo from GitHub.
func (fsl *FsLoader) LoadOneTree(rawPath FilePath) (*MyFolder, error) {
	if smellsLikeGithubCloneArg(string(rawPath)) {
		return CloneAndLoadRepo(fsl, string(rawPath))
	}
	f, err := fsl.LoadFolder(rawPath)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// LoadFolder loads the files at or below a path into memory, returning them
// inside an MyFolder instance. LoadFolder must return a folder, even if its
// argument is a path to a file instead of a path to a folder.
//
// The returned folder's name is computed as follows:
//
//	           path to file | returned folder name | folder contents
//		  ------------------+----------------------+--------------
//		             foo.md |                    . | foo.md
//		           ./foo.md |                    . | foo.md
//		          ../foo.md |            {illegal} | {illegal}
//		  /usr/local/foo.md |           /usr/local | foo.md
//		         bar/foo.md |                  bar | foo.md
//
//		     path to folder | returned folder name | folder contents
//		  ------------------+----------------------+--------------
//		     {empty string} |                    . | {contents of .}
//		                  . |                    . | {contents of .}
//		                foo |                  foo | {contents of foo}
//		              ./foo |                  foo | {contents of foo}
//		             ../foo |            {illegal} | {illegal}
//		     /usr/local/foo |       /usr/local/foo | {contents of foo}
//		            bar/foo |              bar/foo | {contents of foo}
//
// The ".." is disallowed in paths because it screws up making the left nav.
// TODO: consider converting such path to absolute paths.
//
//	The returned folder name might have to be abbreviated.
//
// Files or folders that don't pass the loader's filters are excluded.
// If filtering leaves a folder empty, the folder is discarded.  If nothing
// makes it through, the function returns a nil folder and no error.
//
// If an "OrderingFileName" is found in a folder, it's used to sort the files
// and sub-folders in that folder's in-memory representation. An ordering file
// is just lines of text, one name per line. Ordered files appear first, with
// the remainder in the order imposed by fs.ReadDir.
//
// Any error returned will be from the file system.
func (fsl *FsLoader) LoadFolder(rawPath FilePath) (*MyFolder, error) {
	// If rawPath is empty, cleanPath ends up with "."
	cleanPath := filepath.Clean(string(rawPath))

	// For now, disallow paths that start with upDir, because in the task at
	// hand we want a clear root folder for display. Might allow this later.
	if strings.HasPrefix(cleanPath, string(upDir)) {
		return nil, fmt.Errorf(
			"specify absolute path or something at or below your working folder")
	}

	var (
		err  error
		info os.FileInfo
	)

	info, err = fsl.fs.Stat(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("unable to see folder %q; %w", cleanPath, err)
	}

	if info.IsDir() {
		if err = fsl.IsAllowedFolder(info); err != nil {
			// If user explicitly asked for a disallowed folder, complain.
			// Deeper in, when absorbing folders, they are simply ignored.
			return nil, fmt.Errorf("illegal folder %q; %w", info.Name(), err)
		}
		var fld *MyFolder
		fld, err = fsl.loadFolder(cleanPath)
		if err != nil {
			return nil, err
		}
		if !fld.IsEmpty() {
			fld.name = cleanPath
			return fld, nil
		}
		return nil, nil
	}
	// Load just one file.
	if err = fsl.IsAllowedFile(info); err != nil {
		// If user explicitly asked for a disallowed file, complain.
		// Deeper in, when absorbing folders, they are simply ignored.
		return nil, fmt.Errorf("illegal file %q; %w", info.Name(), err)
	}
	dir, base := DirBase(cleanPath)
	var c []byte
	c, err = fsl.fs.ReadFile(cleanPath)
	if err != nil {
		return nil, err
	}
	return NewFolder(dir).AddFile(NewFile(base, c)), nil
}

// loadFolder loads the folder specified by the path.
// This is the recursive part of the LoadFolder entrypoint.
// The path must point to a folder.
// For example, given a file system like
//
//	/home/bob/
//	  f1.md
//	  games/
//	    doom.md
//
// The argument /home/bob should yield an unnamed, unparented folder containing
// 'f1.md' and the folder 'game' (with 'doom.md' inside 'game').
//
// The same thing is returned if the file system is
//
//	./
//	  f1.md
//	  games/
//	    doom.md
//
// and the argument passed in is simply "." or an empty string.
func (fsl *FsLoader) loadFolder(path string) (*MyFolder, error) {
	var (
		result   MyFolder
		subFld   *MyFolder
		ordering []string
	)
	dirEntries, err := fsl.fs.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf(
			"unable to read folder %q; %w", path, err)
	}
	for i := range dirEntries {
		info := dirEntries[i]
		subPath := filepath.Join(path, info.Name())
		if info.IsDir() {
			if err = fsl.IsAllowedFolder(info); err == nil {
				if subFld, err = fsl.loadFolder(subPath); err != nil {
					return nil, err
				}
				if !subFld.IsEmpty() {
					subFld.name = info.Name()
					result.AddFolder(subFld)
				}
			}
			continue
		}
		if IsOrderingFile(info) {
			// load it and keep it for use at end of function.
			if ordering, err = LoadOrderFile(fsl.fs, subPath); err != nil {
				return nil, err
			}
			continue
		}
		if err = fsl.IsAllowedFile(info); err == nil {
			fi := NewEmptyFile(info.Name())
			fi.content, err = fsl.fs.ReadFile(subPath)
			if err != nil {
				return nil, err
			}
			result.AddFile(fi)
		}
	}
	if result.IsEmpty() {
		return nil, nil
	}
	result.files = ReorderFiles(result.files, ordering)
	result.dirs = ReorderFolders(result.dirs, ordering)
	return &result, nil
}
