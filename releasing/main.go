package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/monopole/mdrip/releasing/internal"
)

// This builds and releases a Go module to GitHub and dockerhub.
// In the old days I'd have hacked this up in unreadable bash code.
//
// Assumptions
//   - The process's working directory is the top of the repo
//     from which code is being released.
//   - The repo has just one Go module (a go.mod file at the top).
//     That's what's being released.
//   - This code doesn't apply tags.  The desired release tag should
//     be applied before running this main to deploy it.
func main() {
	if os.Getenv("GH_TOKEN") == "" {
		log.Fatal("GH_TOKEN not defined, so the gh tool won't work.")
	}
	if len(os.Args) < 2 {
		log.Fatal("Specify the absolute path to the module to build.")
	}
	if len(os.Args) > 2 {
		log.Fatal("Specify only the absolute path to the module to build.")
	}
	dirSrc := os.Args[1]
	if !filepath.IsAbs(dirSrc) {
		log.Fatal(dirSrc + " is not an absolute path.")
	}
	doIt(dirSrc)
}

func doIt(dirSrc string) {
	var (
		tag, commit, dirOut string
		err                 error
		assets              []string
	)
	git := internal.NewGitRunner(internal.DoIt, 30*time.Second)
	if err = git.AssureCleanWorkspace(); err != nil {
		log.Fatal(err)
	}
	tag, commit, err = findTag(git)
	if err != nil {
		if tag != "" {
			slog.Warn("The latest tag " + tag + " doesn't match latest commit.")
		}
		slog.Warn("Define a tag, e.g.:")
		slog.Warn("    tag=v2.0.0-rc10  # i.e., some semver tag")
		slog.Warn("Delete it (if you want to redefine it):")
		slog.Warn("    git push origin :refs/tags/$tag; git tag -d $tag")
		slog.Warn("Create and push it:")
		slog.Warn("    git tag -m \"$tag release\" $tag; git push origin $tag")
		log.Fatal()
	}
	dirOut, err = os.MkdirTemp("", "release-"+filepath.Base(dirSrc)+"-")
	if err != nil {
		log.Fatal(err)
	}
	buildDate := time.Now().UTC()

	ldVars := &internal.LdVars{
		ImportPath: "github.com/monopole/mdrip/v2/internal/provenance",
		Kvs: map[string]string{
			"version":   tag,
			"gitCommit": commit,
			"buildDate": buildDate.Format(time.RFC3339),
		},
	}
	err = buildAndPushDockerImage(dirSrc, dirOut, ldVars)
	if err != nil {
		log.Fatal(err)
	}
	assets, err = buildReleaseAssetsForGitHub(dirSrc, dirOut, ldVars)
	if err != nil {
		log.Fatal(err)
	}
	gh := internal.NewGithubRunner(internal.DoIt, dirSrc, 3*time.Minute)
	if err = gh.Release(tag, assets); err != nil {
		slog.Error(gh.Out())
		log.Fatal(err)
	}
	if gh.Out() != "" {
		slog.Info(gh.Out())
	}
}

func findTag(git *internal.GitRunner) (string, string, error) {
	tag, err := git.GetLatestTag()
	if err != nil {
		return "", "", err
	}
	var tag0, commitHead string
	commitHead, err = git.GetCommitHashOfHead()
	if err != nil {
		return tag, commitHead, err
	}
	tag0, err = git.GetTagAtCommit(commitHead)
	if err != nil {
		return tag, "", err
	}
	if tag != tag0 {
		slog.Warn("         The most recent commit: " + commitHead)
		slog.Warn("  The most recent tag reachable ")
		slog.Warn("         from the latest commit: " + tag0)
		slog.Warn("               The 'latest' tag: " + tag)
		slog.Warn("These two tags don't match; apply a new one?")
		err = fmt.Errorf("tag mismatch")
	}
	return tag, commitHead, err
}

func buildReleaseAssetsForGitHub(
	dirSrc, dirOut string, ldVars *internal.LdVars) ([]string, error) {
	goBuilder := internal.NewGoBuilder(dirSrc, dirOut, ldVars)
	var assetPaths []string
	for _, pair := range []struct {
		myOs   internal.EnumOs
		myArch internal.EnumArch
	}{
		// Add more combinations as desired.
		{myOs: internal.OsLinux, myArch: internal.ArchAmd64},
		{myOs: internal.OsWindows, myArch: internal.ArchAmd64},
		{myOs: internal.OsDarwin, myArch: internal.ArchAmd64},
		{myOs: internal.OsDarwin, myArch: internal.ArchArm64},
	} {
		n, err := goBuilder.Build(pair.myOs, pair.myArch)
		if err != nil {
			return nil, err
		}
		slog.Info("Created " + n)
		assetPaths = append(assetPaths, n)
	}
	return assetPaths, nil
}

func buildAndPushDockerImage(
	dirSrc, dirOut string, ldVars *internal.LdVars) error {
	docker := internal.NewDockerRunner(dirSrc, dirOut, ldVars)
	if err := docker.Login(); err != nil {
		return err
	}
	if err := docker.Build(); err != nil {
		return err
	}
	return docker.Push()
}
