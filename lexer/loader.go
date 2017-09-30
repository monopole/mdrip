package lexer

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/golang/glog"
	"github.com/monopole/mdrip/base"
	"github.com/monopole/mdrip/model"
)

const badLeadingChar = "~.#"

func isDesirableFile(n base.FilePath) bool {
	s, err := os.Stat(string(n))
	if err != nil {
		glog.Info("Stat error on "+s.Name(), err)
		return false
	}
	if s.IsDir() {
		glog.Info("Ignoring NON-file " + s.Name())
		return false
	}
	if !s.Mode().IsRegular() {
		glog.Info("Ignoring irregular file " + s.Name())
		return false
	}
	if filepath.Ext(s.Name()) != ".md" {
		glog.Info("Ignoring non markdown file " + s.Name())
		return false
	}
	base := filepath.Base(s.Name())
	if strings.Index(badLeadingChar, string(base[0])) > -1 {
		glog.Info("Ignoring because bad leading char: " + s.Name())
		return false
	}
	return true
}

func isDesirableDir(n base.FilePath) bool {
	s, err := os.Stat(string(n))
	if err != nil {
		glog.Info("Stat error on "+s.Name(), err)
		return false
	}
	if !s.IsDir() {
		glog.Info("Ignoring NON-dir " + s.Name())
		return false
	}
	// Allow special dir names.
	if s.Name() == "." || s.Name() == "./" || s.Name() == ".." {
		return true
	}
	// Ignore .git, etc.
	if strings.HasPrefix(filepath.Base(s.Name()), ".") {
		glog.Info("Ignoring dot dir " + s.Name())
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
			if err != nil {
				return nil, err
			}
			items = append(items, l)
		} else if isDesirableDir(p) {
			c, err := scanDir(p)
			if err != nil {
				return nil, err
			}
			if c != nil {
				items = append(items, c)
			}
		}
	}
	if len(items) > 0 {
		return model.NewCourse(d, items), nil
	}
	return nil, nil
}

func scanFile(n base.FilePath) (*model.LessonTut, error) {
	contents, err := n.Read()
	if err != nil {
		return nil, err
	}
	return model.NewLessonTutFromBlockParsed(n, Parse(contents)), nil
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

func putReadMeAtTop(x []model.Tutorial) []model.Tutorial {
	return shiftToTop(x, "README")
}

func LoadTutorialFromPath(
	path base.FilePath, nameOverride string) (model.Tutorial, error) {
	if isDesirableFile(path) {
		return scanFile(path)
	}
	if isDesirableDir(path) {
		c, err := scanDir(path)
		if err != nil {
			return nil, err
		}
		if c != nil {
			if len(nameOverride) > 0 {
				path = base.FilePath(nameOverride)
			}
			return model.NewTopCourse(path, putReadMeAtTop(c.Children())), nil
		}
	}
	return nil, errors.New("Unable to grok anything in " + string(path))
}

func LoadTutorialFromPaths(paths []base.FilePath) (model.Tutorial, error) {
	if len(paths) == 0 {
		return nil, errors.New("no paths?")
	}
	if len(paths) == 1 {
		return LoadTutorialFromPath(paths[0], "")
	}
	var items = []model.Tutorial{}
	for _, f := range paths {
		if isDesirableFile(f) {
			l, err := scanFile(f)
			if err != nil {
				return nil, err
			}
			items = append(items, l)
		} else if isDesirableDir(f) {
			c, err := scanDir(f)
			if err != nil {
				return nil, err
			}
			if c != nil {
				items = append(items, c)
			}
		}
	}
	if len(items) > 0 {
		return model.NewTopCourse(base.FilePath(""), putReadMeAtTop(items)), nil
	}
	return nil, errors.New("nothing useful found in paths")
}

func LoadTutorialFromGitHub(url string) (model.Tutorial, error) {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		glog.Error("No git on path", err)
		return nil, err
	}
	tmpDir, err := ioutil.TempDir("", "mdrip-unpack-")
	if err != nil {
		glog.Error("Unable to create tmp dir", err)
		return nil, err
	}
	glog.Info("Using " + gitPath + " to clone to " + tmpDir)
	defer os.RemoveAll(tmpDir)
	cmd := exec.Command(gitPath, "clone", url, tmpDir)
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		glog.Error("git clone failure", err)
		return nil, err
	}
	return LoadTutorialFromPath(base.FilePath(tmpDir), url)
}
