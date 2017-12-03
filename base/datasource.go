package base

import (
	"errors"
	"path/filepath"

	"strings"
)

type DataSource struct {
	raw      string
	repoName string
	relPath  string
	absPath  string
}

func (d *DataSource) IsGithub() bool {
	return len(d.repoName) > 0
}

func (d *DataSource) Display() string {
	if d.IsGithub() {
		result := "gh:" + d.repoName
		if len(d.relPath) > 0 {
			return result + "/" + d.relPath
		}
		return result
	}
	return d.raw
}

func (d *DataSource) Href() string {
	if d.IsGithub() {
		result := "https://github.com/" + d.repoName
		if len(d.relPath) > 0 {
			return result + "/blob/master/" + d.relPath
		}
		return result
	}
	return "file://" + d.absPath
}

// Using https instead of ssh so no need for keys
// (works only with public repos obviously).
func (d *DataSource) GithubCloneArg() string {
	return "https://github.com/" + d.repoName + ".git"
}

func (d *DataSource) RelPath() FilePath {
	return FilePath(d.relPath)
}

func (d *DataSource) AbsPath() FilePath {
	return FilePath(d.absPath)
}

func (d *DataSource) SetAbsPath(arg string) {
	d.absPath = arg
}

func (d *DataSource) Raw() string {
	return d.raw
}

func NewDataSource(arg string) (*DataSource, error) {
	n := strings.TrimSpace(arg)
	if len(n) < 1 {
		return nil, errors.New(
			"Need data source - file name, directory name, or github clone url.")
	}
	if smellsLikeGithubCloneArg(arg) {
		repoName, path, err := extractGithubRepoName(arg)
		if err != nil {
			return nil, err
		}
		return &DataSource{arg, repoName, path, ""}, nil
	}
	path, err := filepath.Abs(arg)
	if err != nil {
		return nil, errors.New(
			"Unable to resolve absolute path of " + arg)
	}
	return &DataSource{arg, "", arg, path}, nil
}

func smellsLikeGithubCloneArg(arg string) bool {
	arg = strings.ToLower(arg)
	return strings.HasPrefix(arg, "gh:") ||
		strings.HasPrefix(arg, "github.com") ||
		strings.HasPrefix(arg, "git@github.com:") ||
		strings.Index(arg, "github.com/") > -1
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
