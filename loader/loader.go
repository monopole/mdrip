package loader

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/golang/glog"
	"github.com/monopole/mdrip/base"
	"github.com/monopole/mdrip/lexer"
	"github.com/monopole/mdrip/model"
	"github.com/pkg/errors"
)

const (
	badLeadingChar = "~.#"
)

func isOrderFile(n base.FilePath) bool {
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

func isDesirableFile(n base.FilePath) bool {
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
	if strings.Index(badLeadingChar, string(base[0])) > -1 {
		return false
	}
	return true
}

func isDesirableDir(n base.FilePath) bool {
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

func scanDir(d base.FilePath) (model.Tutorial, error) {
	files, err := d.ReadDir()
	if err != nil {
		return BadLoad(d), err
	}
	var items = []model.Tutorial{}
	var ordering = []string{}
	for _, f := range files {
		p := d.Join(f)
		if isDesirableFile(p) {
			l, err := scanFile(p)
			if err == nil {
				items = append(items, l)
			}
			continue
		}
		if isDesirableDir(p) {
			c, err := scanDir(p)
			if err == nil {
				items = append(items, c)
			}
			continue
		}
		if isOrderFile(p) {
			contents, err := p.Read()
			if err == nil {
				ordering = strings.Split(contents, "\n")
			}
		}
	}
	if len(items) == 0 {
		return nil, errors.New("no content in directory " + string(d))
	}
	return model.NewCourse(d, reorder(items, ordering)), nil
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

// A tutorial complaining about its data source.
func BadLoad(n base.FilePath) model.Tutorial {
	result := model.NewMdContent()
	result.AddBlockParsed(model.NewProseOnlyBlock(base.MdProse(
		"## Unable to load data from _" + string(n) + "_\n")))
	return model.NewLessonTutFromMdContent(n, result)
}

func shiftToTop(x []model.Tutorial, top string) []model.Tutorial {
	result := []model.Tutorial{}
	other := []model.Tutorial{}
	for _, f := range x {
		if f.Name() == top {
			result = append(result, f)
		} else {
			other = append(other, f)
		}
	}
	return append(result, other...)
}

// reorder tutorial array in some fashion
func reorder(x []model.Tutorial, ordering []string) []model.Tutorial {
	for i := len(ordering) - 1; i >= 0; i-- {
		x = shiftToTop(x, ordering[i])
	}
	return shiftToTop(x, "README")
}

type Loader struct {
	ds *base.DataSet
}

func (l *Loader) DataSet() *base.DataSet {
	return l.ds
}

func NewLoader(ds *base.DataSet) *Loader {
	return &Loader{ds}
}

func (l *Loader) SmellsLikeGithub() bool {
	if l.ds.N() != 1 {
		return false
	}
	return l.ds.FirstArg().IsGithub()
}

func (l *Loader) Load() (model.Tutorial, error) {
	if l.ds.N() == 1 {
		if l.ds.FirstArg().IsGithub() {
			return loadTutorialFromGitHub(l.ds.FirstArg())
		}
		return loadTutorialFromPath(l.ds.FirstArg())
	}
	// yuck.
	return loadTutorialFromPaths(l.ds.FirstArg(), l.ds.AsPaths())
}

func loadTutorialFromPath(source *base.DataSource) (model.Tutorial, error) {
	if isDesirableFile(source.AbsPath()) {
		return scanFile(source.AbsPath())
	}
	if !isDesirableDir(source.AbsPath()) {
		return nil, errors.New("nothing found at " + string(source.AbsPath()))
	}
	glog.Infof("Loading %s from path %s\n", source.Display(), source.AbsPath())

	c, err := scanDir(source.AbsPath())
	if err != nil {
		return BadLoad(source.AbsPath()), err
	}
	return model.NewTopCourse(source.Display(), source.AbsPath(), c.Children()), nil
}

func loadTutorialFromPaths(source *base.DataSource, paths []base.FilePath) (model.Tutorial, error) {
	var items = []model.Tutorial{}
	for _, f := range paths {
		if isDesirableFile(f) {
			l, err := scanFile(f)
			if err == nil {
				items = append(items, l)
			}
			continue
		}
		if isDesirableDir(f) {
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
	glog.Infof("Deleted " + tmpDir)
}

func loadTutorialFromGitHub(source *base.DataSource) (model.Tutorial, error) {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return BadLoad(base.FilePath(source.Raw())),
			errors.Wrap(err, "maybe no git on path")
	}
	tmpDir, err := ioutil.TempDir("", "mdrip-git-")
	if err != nil {
		return BadLoad(base.FilePath(source.Raw())),
			errors.Wrap(err, "unable to create tmp dir")
	}
	glog.Infof("Cloning to %s ...\n", tmpDir)
	defer cleanUp(tmpDir)
	cmd := exec.Command(gitPath, "clone", source.GithubCloneArg(), tmpDir)
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return BadLoad(base.FilePath(source.Raw())),
			errors.Wrap(err, "git clone failure")
	}
	glog.Info("Clone complete.")
	fullPath := tmpDir
	if len(source.RelPath()) > 0 {
		fullPath = filepath.Join(fullPath, string(source.RelPath()))
	}
	source.SetAbsPath(fullPath)
	return loadTutorialFromPath(source)
}
