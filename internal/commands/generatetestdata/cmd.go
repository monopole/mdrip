package generatetestdata

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/monopole/mdrip/v2/internal/loader"
	"github.com/monopole/mdrip/v2/internal/utils"
	"github.com/monopole/mdrip/v2/internal/web/app/widget/testutil"
	"github.com/spf13/cobra"
)

const (
	defaultDirName = "testdata"
	cmdName        = "generatetestdata"
	shortHelp      = "Create a disposable folder containing markdown for use in tests"
)

func NewCommand() *cobra.Command {
	flags := myFlags{}
	c := &cobra.Command{
		Use:   cmdName,
		Short: shortHelp,
		Long: shortHelp + `

The folder name provided should not exist; this command wants to write
to an empty folder to create known state for testing.

The default folder name is '` + defaultDirName + `'.

Having ` + utils.PgmName + ` contain a means to generate test data
removes the need to download anything other than the binary to perform
tests on a particular os/architecture.
`,
		Example: utils.PgmName + " " + cmdName + " {nameOfNewFolder}",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return fmt.Errorf(
					"specify only the name of the folder to create")
			}
			path := defaultDirName
			if len(args) == 1 {
				path = args[0]
			}
			stat, err := utils.PathStatus(path)
			if err != nil {
				return fmt.Errorf("path %q has a problem; %w", path, err)
			}
			if stat == utils.PathIsAFile {
				return fmt.Errorf("%q is a file; not removing", path)
			}
			// TODO: Should we be paranoid?
			//  if filepath.IsAbs(path) {
			// 	  return fmt.Errorf("not allowing absolute paths at this time")
			//  }
			if stat == utils.PathIsAFolder {
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
			slog.Debug("Created folder " + path)
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

type myFlags struct {
	overwrite bool
}

type treeWriter struct{}

func mkDirOrDie(fp loader.FilePath) {
	if err := os.MkdirAll(string(fp), os.ModePerm); err != nil {
		slog.Error("mkdir failure", "err", err)
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
		slog.Error("unable to visit file", "file", fi.Path(), "err", err)
	}
}

func (v *treeWriter) Error() error { return nil }
