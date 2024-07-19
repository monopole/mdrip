package gentestdata

import (
	"errors"
	"fmt"
	"github.com/monopole/mdrip/v2/internal/loader"
	"github.com/monopole/mdrip/v2/internal/utils"
	"github.com/monopole/mdrip/v2/internal/web/app/widget/testutil"
	"github.com/spf13/cobra"
	"log"
	"log/slog"
	"os"
)

const (
	defaultDirName = "testdata"
	cmdName        = "gen" + defaultDirName
	shortHelp      = "Creates a disposable folder containing markdown files for use in tests."
)

func NewCommand() *cobra.Command {
	flags := myFlags{}
	c := &cobra.Command{
		Use:   cmdName,
		Short: shortHelp,
		Long: shortHelp + `
The folder should not exist; this command wants to write
to an empty folder to create known state for testing.
`,
		Example: utils.PgmName + " " + cmdName + " {path/to/folder}",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return fmt.Errorf("specify only one argument, the path to a folder")
			}
			path := defaultDirName
			if len(args) == 1 {
				path = args[0]
			}
			stat, err := pathStatus(path)
			if err != nil {
				return fmt.Errorf("path %q has a problem; %w", path, err)
			}
			if stat == pathIsAFile {
				return fmt.Errorf("%q is a file; not removing", path)
			}
			// TODO: Be paranoid?
			//  if filepath.IsAbs(path) {
			// 	  return fmt.Errorf("not allowing absolute paths at this time")
			//  }
			if stat == pathIsAFolder {
				if !flags.overwrite {
					return fmt.Errorf("folder %q exists; not overwriting", path)
				}
				if err = os.RemoveAll(path); err != nil {
					return fmt.Errorf("trouble deleting %q; %w", path, err)
				}
			}
			// Getting here means we can create the folder
			// and write files into it.
			f := testutil.MakeNamedFolderTreeOfMarkdown(loader.NewFolder(path))
			f.Accept(&treeWriter{})
			slog.Info("Created folder " + path)
			return nil
		},
	}
	c.Flags().BoolVar(
		&flags.overwrite,
		"overwrite",
		false,
		"Overwrite the given folder if it exists.")
	return c
}

type pathCase int

const (
	pathInUnknownState pathCase = iota
	pathIsAFile
	pathIsAFolder
	pathDoesNotExist
)

func pathStatus(path string) (pathCase, error) {
	fi, err := os.Stat(path)
	if err == nil {
		// path exists!
		if fi.IsDir() {
			return pathIsAFolder, nil
		}
		return pathIsAFile, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return pathDoesNotExist, nil
	}
	// File may or may not exist, depends on the error.
	return pathInUnknownState, err
}

type myFlags struct {
	overwrite bool
}

type treeWriter struct{}

func mkDirOrDie(fp loader.FilePath) {
	if err := os.MkdirAll(string(fp), os.ModePerm); err != nil {
		log.Fatal(err)
	}
}

func (v *treeWriter) VisitTopFolder(fl *loader.MyTopFolder) {
	mkDirOrDie(fl.Path())
	fl.VisitChildren(v)
}

func (v *treeWriter) VisitFolder(fl *loader.MyFolder) {
	mkDirOrDie(fl.Path())
	fl.VisitChildren(v)
}

func (v *treeWriter) VisitFile(fi *loader.MyFile) {
	if err := os.WriteFile(string(fi.Path()), fi.C(), 0666); err != nil {
		log.Fatal(err)
	}
}

func (v *treeWriter) Error() error { return nil }
