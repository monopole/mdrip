package tutorial

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang/glog"
	"github.com/monopole/mdrip/model"
)

const badLeadingChar = "~.#"

func isDesirableFile(n model.FilePath) bool {
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

func isDesirableDir(n model.FilePath) bool {
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

func scanDir(d model.FilePath) (*Course, error) {
	files, err := d.ReadDir()
	if err != nil {
		return nil, err
	}
	var items = []Tutorial{}
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
		return NewCourse(d, items), nil
	}
	return nil, nil
}

func scanFile(n model.FilePath) (*Lesson, error) {
	contents, err := n.Read()
	if err != nil {
		return nil, err
	}
	return NewLesson(n, contents), nil
}

func LoadTutorialFromPath(path model.FilePath) (Tutorial, error) {
	if isDesirableFile(path) {
		return scanFile(path)
	}
	if isDesirableDir(path) {
		c, err := scanDir(path)
		if err != nil {
			return nil, err
		}
		if c != nil {
			return NewTopCourse(path, c.children), nil
		}
	}
	return nil, errors.New("cannot load from " + string(path))
}

func LoadTutorialFromPaths(paths []model.FilePath) (Tutorial, error) {
	if len(paths) == 0 {
		return nil, errors.New("no paths?")
	}
	if len(paths) == 1 {
		return LoadTutorialFromPath(paths[0])
	}
	var items = []Tutorial{}
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
		return NewTopCourse(model.FilePath(""), items), nil
	}
	return nil, errors.New("nothing useful found in paths")
}

// Build program code from blocks extracted from a tutorial.
func NewProgramFromTutorial(l model.Label, t Tutorial) *model.Program {
	v := NewScriptExtractor(l)
	t.Accept(v)
	return model.NewProgram(l, v.Scripts())
}

// Build program code from blocks extracted from markdown files.
func NewProgramFromPaths(l model.Label, paths []model.FilePath) (*model.Program, error) {
	t, err := LoadTutorialFromPaths(paths)
	if err != nil {
		return nil, err
	}
	return NewProgramFromTutorial(l, t), nil
}
