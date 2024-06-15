package loader

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const dotGit = ".git"

// smellsLikeGithubCloneArg returns true if the argument seems
// like it could be GitHub url or `git clone` argument.
func smellsLikeGithubCloneArg(arg string) bool {
	arg = strings.ToLower(arg)
	return strings.HasPrefix(arg, "gh:") ||
		strings.HasPrefix(arg, "git@github.com:") ||
		strings.HasPrefix(arg, "https://github.com/")
}

// CloneAndLoadRepo clones a repo locally and loads it.
// The FsLoader should be injected with a real file system,
// since the git command line used here clones to real disk.
func CloneAndLoadRepo(fsl *FsLoader, arg string) (*MyFolder, error) {
	d, n := splitDomainAndRemainder(arg)
	r, p, err := splitRepoAndPath(n)
	if err != nil {
		return nil, err
	}
	var (
		tmpDir string
		fld    *MyFolder
	)
	tmpDir, err = cloneRepo(d, r)
	defer os.RemoveAll(tmpDir)
	if err != nil {
		return nil, err
	}
	fld, err = fsl.LoadFolder(FilePath(filepath.Join(tmpDir, p)))
	if err != nil {
		return nil, err
	}
	if fld.NumFiles() == 1 && fld.NumFolders() == 0 {
		p = strings.TrimSuffix(p, fld.files[0].name)
		p = strings.TrimSuffix(p, string(RootSlash))
	}
	if p == "" {
		fld.name = d + r
	} else {
		fld.name = d + r + string(RootSlash) + p
	}
	return fld, nil
}

// splitRepoAndPath parses strings like monopole/mdrip.git/somepath or
// monopole/mdrip, splitting the repository name
// and the path inside the repository.
func splitRepoAndPath(n string) (string, string, error) {
	if i := strings.Index(n, dotGit); i > 0 {
		r := n[:i]
		after := n[i+len(dotGit):]
		if len(after) > 1 {
			if !strings.HasPrefix(after, "/") {
				return "", "", fmt.Errorf("no path separator in github spec")
			}
			return r, after[1:], nil
		}
		return r, "", nil
	}
	i := strings.Index(n, "/")
	if i < 1 {
		// expect ORGANIZATION/REPONAME
		return "", "", fmt.Errorf("no org/repo separator in github spec")
	}
	// Now look for a second path separator
	j := strings.Index(n[i+1:], "/")
	if j < 0 {
		// No path, so show entire repo.
		return n, "", nil
	}
	j += i + 1
	return n[:j], n[j+1:], nil
}

func splitDomainAndRemainder(raw string) (string, string) {
	n := strings.ToLower(raw)
	{
		p := "gh:"
		if n[:len(p)] == p {
			return "git@github.com:", raw[len(p):]
		}
	}
	// try any domain-like thing
	for _, p := range []string{".com/", ".com:"} {
		if i := strings.Index(n, p); i > 0 {
			return n[:i+len(p)], raw[i+len(p):]
		}
	}
	// err?
	return raw, ""
}

func cloneRepo(domain, repoName string) (string, error) {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return "", fmt.Errorf("maybe no git program? (%w)", err)
	}
	tmpDir, err := os.MkdirTemp("", "mdrip-git-")
	if err != nil {
		return "", fmt.Errorf("unable to create tmp dir (%w)", err)
	}
	slog.Info("Cloning", "tmpDir", tmpDir, "domain", domain, "repoName", repoName)
	cmd := exec.Command(
		gitPath, "clone", domain+repoName+dotGit, tmpDir)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err = cmd.Run(); err != nil {
		return "", fmt.Errorf("git clone failure (%w)", err)
	}
	slog.Info("Clone complete.")
	return tmpDir, nil
}
