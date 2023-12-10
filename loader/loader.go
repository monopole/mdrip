package loader

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/monopole/mdrip/base"
	"github.com/monopole/mdrip/lexer"
	"github.com/monopole/mdrip/model"
	"github.com/pkg/errors"
)

func scanDir(dir base.FilePath) (model.Tutorial, error) {
	dirEntries, err := dir.ReadDir()
	if err != nil {
		return BadLoad(dir), err
	}
	var (
		items    []model.Tutorial
		ordering []string
	)
	for _, f := range dirEntries {
		p := dir.Join(f)
		if p.IsDesirableFile() {
			if tut, er := scanFile(p); er == nil {
				items = append(items, tut)
			}
			continue
		}
		if p.IsDesirableDir() {
			if tut, er := scanDir(p); er == nil {
				items = append(items, tut)
			}
			continue
		}
		if p.IsOrderFile() {
			if contents, er := p.Read(); er == nil {
				ordering = strings.Split(contents, "\n")
			}
		}
	}
	if len(items) == 0 {
		return nil, errors.New("no content in directory " + string(dir))
	}
	return model.NewCourse(dir, reorder(items, ordering)), nil
}

func scanFile(n base.FilePath) (model.Tutorial, error) {
	contents, err := n.Read()
	if err != nil {
		return BadLoad(n), err
	}
	md := lexer.Parse(contents)
	if len(md.Blocks) < 1 {
		return BadLoad(n), errors.New("no content in " + string(n))
	}
	return model.NewLessonTutFromMdContent(n, md), nil
}

// BadLoad returns a fake tutorial complaining about its data source.
// For use with a web browser, to make the problem obvious.
func BadLoad(n base.FilePath) model.Tutorial {
	result := model.NewMdContent()
	result.AddBlockParsed(model.NewProseOnlyBlock(base.MdProse(
		"## Unable to load data from _" + string(n) + "_\n")))
	return model.NewLessonTutFromMdContent(n, result)
}

// Loader loads a dataset.
type Loader struct {
	ds *base.DataSet
}

// DataSet that the Loader will load.
func (l *Loader) DataSet() *base.DataSet {
	return l.ds
}

// NewLoader returns a Loader for the given DataSet.
func NewLoader(ds *base.DataSet) *Loader {
	return &Loader{ds}
}

// SmellsLikeGithub is true if the DataSet smells like github.
func (l *Loader) SmellsLikeGithub() bool {
	if l.ds.Size() != 1 {
		return false
	}
	return l.ds.FirstArg().IsGithub()
}

// Load loads the DataSet into a Tutorial.
func (l *Loader) Load() (model.Tutorial, error) {
	if l.ds.Size() == 1 {
		if l.ds.FirstArg().IsGithub() {
			return loadTutorialFromGitHub(l.ds.FirstArg())
		}
		return loadTutorialFromPath(l.ds.FirstArg())
	}
	// yuck.
	return loadTutorialFromPaths(l.ds.FirstArg(), l.ds.AsPaths())
}

func loadTutorialFromPath(source *base.DataSource) (model.Tutorial, error) {
	if source.AbsPath().IsDesirableFile() {
		return scanFile(source.AbsPath())
	}
	if !source.AbsPath().IsDesirableDir() {
		return nil, errors.New("nothing found at " + string(source.AbsPath()))
	}
	slog.Info(fmt.Sprintf("Loading %s from path %s\n", source.Display(), source.AbsPath()))

	c, err := scanDir(source.AbsPath())
	if err != nil {
		return BadLoad(source.AbsPath()), err
	}
	return model.NewTopCourse(source.Display(), source.AbsPath(), c.Children()), nil
}

func loadTutorialFromPaths(source *base.DataSource, paths []base.FilePath) (model.Tutorial, error) {
	var items []model.Tutorial
	for _, f := range paths {
		if f.IsDesirableFile() {
			l, err := scanFile(f)
			if err == nil {
				items = append(items, l)
			}
			continue
		}
		if f.IsDesirableDir() {
			c, err := scanDir(f)
			if err == nil {
				items = append(items, c)
			}
			continue
		}
	}
	if len(items) == 0 {
		return BadLoad(paths[0]), errors.New("nothing useful found in paths")
	}
	return model.NewTopCourse(source.Display(), source.AbsPath(), items), nil
}

func cleanUp(tmpDir string) {
	os.RemoveAll(tmpDir)
	slog.Info("Deleted " + tmpDir)
}

func loadTutorialFromGitHub(source *base.DataSource) (model.Tutorial, error) {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return BadLoad(base.FilePath(source.Raw())),
			errors.Wrap(err, "maybe no git on path")
	}
	tmpDir, err := os.MkdirTemp("", "mdrip-git-")
	if err != nil {
		return BadLoad(base.FilePath(source.Raw())),
			errors.Wrap(err, "unable to create tmp dir")
	}
	slog.Info("Cloning to " + tmpDir)
	defer cleanUp(tmpDir)
	cmd := exec.Command(gitPath, "clone", source.GithubCloneArg(), tmpDir)
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return BadLoad(base.FilePath(source.Raw())),
			errors.Wrap(err, "git clone failure")
	}
	slog.Info("Clone complete.")
	fullPath := tmpDir
	if len(source.RelPath()) > 0 {
		fullPath = filepath.Join(fullPath, string(source.RelPath()))
	}
	source.SetAbsPath(fullPath)
	return loadTutorialFromPath(source)
}
