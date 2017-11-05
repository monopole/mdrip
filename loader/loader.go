package loader

import (
	"bytes"
	"fmt"
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

func scanFile(n base.FilePath) (model.Tutorial, error) {
	contents, err := n.Read()
	if err != nil {
		return BadLoad(n), err
	}
	parsed := lexer.Parse(contents)
	if len(parsed) < 1 {
		return BadLoad(n), errors.New("no content in " + string(n))
	}
	return model.NewLessonTutFromBlockParsed(n, parsed), nil
}

// A tutorial complaining about its data source.
func BadLoad(n base.FilePath) model.Tutorial {
	blockParsed := model.NewProseOnlyBlock(base.MdProse(
		"## Unable to load data from _" + string(n) + "_\n"))
	blocks := []*model.BlockParsed{blockParsed}
	return model.NewLessonTutFromBlockParsed(n, blocks)
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
func reorder(x []model.Tutorial) []model.Tutorial {
	return shiftToTop(x, "README")
}

type Loader struct {
	ds *base.DataSource
}

func (l *Loader) Source() string {
	return l.ds.String()
}

func NewLoader(ds *base.DataSource) *Loader {
	return &Loader{ds}
}

func smellsLikeGithubCloneArg(arg string) bool {
	arg = strings.ToLower(arg)
	return strings.HasPrefix(arg, "gh:") ||
		strings.HasPrefix(arg, "github.com") ||
		strings.HasPrefix(arg, "git@github.com:") ||
		strings.Index(arg, "github.com/") > -1
}

// buildGithubCloneArg builds an arg for 'git clone' from a repo name.
// Using https instead of ssh so no need for keys
// (works only with public repos obviously).
func buildGithubCloneArg(repoName string) string {
	return "https://github.com/" + repoName + ".git"
}

// From strings like git@github.com:monopole/mdrip.git or
// https://github.com/monopole/mdrip, extract github.com.
func extractGithubRepoName(n string) (string, string, error) {
	for _, p := range []string{
		// Order matters here.
		"gh:", "https://", "http://", "git@", "github.com:", "github.com/"} {
		if strings.ToLower(n[:len(p)]) == p {
			n = n[len(p):]
		}
	}
	if strings.HasSuffix(n, ".git") {
		n = n[0 : len(n)-len(".git")]
	}
	i := strings.Index(n, string(filepath.Separator))
	if i < 1 {
		return "", "", errors.New("No separator.")
	}
	j := strings.Index(n[i+1:], string(filepath.Separator))
	if j < 0 {
		// No path, so show entire repo.
		return n, "", nil
	}
	j += i + 1
	return n[:j], n[j+1:], nil
}

func (l *Loader) SmellsLikeGithub() bool {
	if l.ds.N() != 1 {
		return false
	}
	return smellsLikeGithubCloneArg(l.ds.FirstArg())
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
	glog.Infof("Loading %s from path %s\n", name, path)

	c, err := scanDir(path)
	if err != nil {
		return BadLoad(path), err
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
		return BadLoad(paths[0]), errors.New("nothing useful found in paths")
	}
	return model.NewTopCourse(name, base.FilePath(name), reorder(items)), nil
}

func loadTutorialFromGitHub(url string) (model.Tutorial, error) {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return BadLoad(base.FilePath(url)),
			errors.Wrap(err, "maybe no git on path")
	}
	tmpDir, err := ioutil.TempDir("", "mdrip-git-")
	if err != nil {
		return BadLoad(base.FilePath(url)),
			errors.Wrap(err, "unable to create tmp dir")
	}
	glog.Info("Using " + gitPath + " to clone to " + tmpDir)
	defer os.RemoveAll(tmpDir)
	repoName, path, err := extractGithubRepoName(url)
	cmd := exec.Command(
		gitPath, "clone", buildGithubCloneArg(repoName), tmpDir)
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return BadLoad(base.FilePath(url)),
			errors.Wrap(err, "git clone failure")
	}
	fullPath := tmpDir
	if len(path) > 0 {
		fullPath = filepath.Join(fullPath, path)
	}
	return loadTutorialFromPath("gh:"+repoName, base.FilePath(fullPath))
}
