package lexer

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"fmt"
	"github.com/golang/glog"
	"github.com/monopole/mdrip/base"
	"github.com/monopole/mdrip/model"
)

const (
	badLeadingChar = "~.#"
)

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

func scanDir(d base.FilePath) (*model.Course, error) {
	files, err := d.ReadDir()
	if err != nil {
		return nil, err
	}
	var items = []model.Tutorial{}
	for _, f := range files {
		p := d.Join(f)
		if isDesirableFile(p) {
			l, err := scanFile(p)
			if err == nil {
				items = append(items, l)
			}
		} else if isDesirableDir(p) {
			c, err := scanDir(p)
			if err == nil {
				items = append(items, c)
			}
		}
	}
	if len(items) == 0 {
		return nil, errors.New("no content in directory " + string(d))
	}
	return model.NewCourse(d, items), nil
}

func scanFile(n base.FilePath) (*model.LessonTut, error) {
	contents, err := n.Read()
	if err != nil {
		return nil, err
	}
	parsed := Parse(contents)
	if len(parsed) < 1 {
		return nil, errors.New("no content in " + string(n))
	}
	return model.NewLessonTutFromBlockParsed(n, parsed), nil
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

func reorder(x []model.Tutorial) []model.Tutorial {
	return shiftToTop(x, "README")
}

type Loader struct {
	ds *base.DataSource
}

func NewLoader(ds *base.DataSource) *Loader {
	return &Loader{ds}
}

func smellsLikeGithubCloneArg(arg string) bool {
	return strings.HasPrefix(arg, "gh:") ||
		strings.HasPrefix(arg, "GH:") ||
		strings.HasPrefix(arg, "github.com") ||
		strings.HasPrefix(arg, "git@github.com:") ||
		strings.Index(arg, "github.com/") > -1
}

// buildGithubCloneArg builds an arg for 'git clone' from a repo name.
func buildGithubCloneArg(repoName string) string {
	return "git@github.com:" + repoName + ".git"
}

// From strings like git@github.com:monopole/mdrip.git or
// https://github.com/monopole/mdrip, extract github.com.
func extractGithubRepoName(n string) string {
	remove := "GH:"
	if strings.HasPrefix(n, remove) {
		n = n[len(remove):]
	}
	remove = "gh:"
	if strings.HasPrefix(n, remove) {
		n = n[len(remove):]
	}
	remove = "https://"
	if strings.HasPrefix(n, remove) {
		n = n[len(remove):]
	}
	remove = "http://"
	if strings.HasPrefix(n, remove) {
		n = n[len(remove):]
	}
	remove = "git@"
	if strings.HasPrefix(n, remove) {
		n = n[len(remove):]
	}
	remove = "github.com:"
	if strings.HasPrefix(n, remove) {
		n = n[len(remove):]
	}
	remove = "github.com/"
	if strings.HasPrefix(n, remove) {
		n = n[len(remove):]
	}
	if strings.HasSuffix(n, ".git") {
		n = n[0 : len(n)-len(".git")]
	}
	return n
}

func (l *Loader) Load() (model.Tutorial, error) {
	if l.ds.N() == 1 {
		if smellsLikeGithubCloneArg(l.ds.FirstArg()) {
			return loadTutorialFromGitHub(l.ds.FirstArg())
		}
		p := base.FilePath(l.ds.FirstArg())
		return loadTutorialFromPath(p.Base(), p)
	}
	name := fmt.Sprintf("(%d paths)", l.ds.N())
	return loadTutorialFromPaths(name, l.ds.AsPaths())
}

func loadTutorialFromPath(name string, path base.FilePath) (model.Tutorial, error) {
	if isDesirableFile(path) {
		return scanFile(path)
	}
	if !isDesirableDir(path) {
		return nil, errors.New("nothing found at file path " + string(path))
	}
	c, err := scanDir(path)
	if err != nil {
		return nil, err
	}
	return model.NewTopCourse(name, path, reorder(c.Children())), nil
}

func loadTutorialFromPaths(name string, paths []base.FilePath) (model.Tutorial, error) {
	var items = []model.Tutorial{}
	for _, f := range paths {
		if isDesirableFile(f) {
			l, err := scanFile(f)
			if err == nil {
				items = append(items, l)
			}
		} else if isDesirableDir(f) {
			c, err := scanDir(f)
			if err == nil {
				items = append(items, c)
			}
		}
	}
	if len(items) == 0 {
		return nil, errors.New("nothing useful found in paths")
	}
	return model.NewTopCourse(name, base.FilePath(name), reorder(items)), nil
}

func loadTutorialFromGitHub(url string) (model.Tutorial, error) {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		glog.Error("No git on path", err)
		return nil, err
	}
	tmpDir, err := ioutil.TempDir("", "mdrip-git-")
	if err != nil {
		glog.Error("Unable to create tmp dir", err)
		return nil, err
	}
	glog.Info("Using " + gitPath + " to clone to " + tmpDir)
	defer os.RemoveAll(tmpDir)
	repoName := extractGithubRepoName(url)
	cmd := exec.Command(
		gitPath, "clone", buildGithubCloneArg(repoName), tmpDir)
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		glog.Error("git clone failure ", err)
		return nil, err
	}
	return loadTutorialFromPath("gh:"+repoName, base.FilePath(tmpDir))
}
