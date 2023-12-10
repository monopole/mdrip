package base

import (
	"errors"
	"path/filepath"

	"strings"
)

// DataSource is where markdown came from.
type DataSource struct {
	// The raw, original specification of the data source.
	raw string
	// The GitHub repository name, if any.
	repoName string
	// The path to the file or directory of interest, relative
	// to the GitHub repo.  If empty, then the whole repo is used.
	relPath string
	// The absolute path to the local directory.
	absPath string
}

// IsGithub is true if the datasource was GitHub.
func (d *DataSource) IsGithub() bool {
	return len(d.repoName) > 0
}

// Display is a string intended for display.
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

// Href returns a browser url compatible form of the datasource.
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

// GithubCloneArg returns the data source in a form suitable for `git clone`.
func (d *DataSource) GithubCloneArg() string {
	return "https://github.com/" + d.repoName + ".git"
}

// RelPath is the relative file system path of the datasource.
func (d *DataSource) RelPath() FilePath {
	return FilePath(d.relPath)
}

// AbsPath is the absolute file system path of the datasource.
func (d *DataSource) AbsPath() FilePath {
	return FilePath(d.absPath)
}

// SetAbsPath changes the path.
func (d *DataSource) SetAbsPath(arg string) {
	d.absPath = arg
}

// Raw form of the datasource.
func (d *DataSource) Raw() string {
	return d.raw
}

// NewDataSource ctor.
func NewDataSource(arg string) (*DataSource, error) {
	n := strings.TrimSpace(arg)
	if len(n) < 1 {
		return nil, errors.New(
			"need data source - file name, directory name, or github clone url")
	}
	if SmellsLikeGithubCloneArg(arg) {
		repoName, path, err := ExtractGithubRepoName(arg)
		if err != nil {
			return nil, err
		}
		return &DataSource{raw: arg, repoName: repoName, relPath: path}, nil
	}
	path, err := filepath.Abs(arg)
	if err != nil {
		return nil, errors.New(
			"unable to resolve absolute path of " + arg)
	}
	return &DataSource{raw: arg, relPath: arg, absPath: path}, nil
}

// SmellsLikeGithubCloneArg returns true if the argument seems
// like it could be GitHub url or `git clone` argument.
func SmellsLikeGithubCloneArg(arg string) bool {
	arg = strings.ToLower(arg)
	return strings.HasPrefix(arg, "gh:") ||
		strings.HasPrefix(arg, "github.com") ||
		strings.HasPrefix(arg, "git@github.com:") ||
		strings.Index(arg, "github.com/") > -1
}

// ExtractGithubRepoName parses strings like git@github.com:monopole/mdrip.git or
// https://github.com/monopole/mdrip, extracting the repository name
// and the path inside the repository.
func ExtractGithubRepoName(n string) (string, string, error) {
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
	i := strings.Index(n, "/")
	if i < 1 {
		return "", "", errors.New("no separator")
	}
	j := strings.Index(n[i+1:], "/")
	if j < 0 {
		// No path, so show entire repo.
		return n, "", nil
	}
	j += i + 1
	return n[:j], n[j+1:], nil
}
