package main

import (
	"fmt"
	"github.com/monopole/mdrip/releasing/internal"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

const envSrcDir = "MODULE_DIR"

func main() {
	git := internal.NewGitRunner(internal.DoIt, 30*time.Second)
	tag, commit, err := findTag(git)
	//if err != nil {
	//	log.Fatal(err)
	//}
	dirSrc := os.Getenv(envSrcDir)
	if dirSrc == "" {
		log.Fatal("You must define the env var " + envSrcDir)
	}
	dirOut, err := os.MkdirTemp("", "release-"+filepath.Base(dirSrc))
	if err != nil {
		log.Fatal(err)
	}
	err = buildBinaries(dirSrc, dirOut, tag, commit)
	if err != nil {
		log.Fatal(err)
	}
}

func findTag(git *internal.GitRunner) (string, string, error) {
	tag, err := git.GetLatestTag()
	if err != nil {
		return "", "", err
	}
	var commitAtTag, commitLatest string
	commitAtTag, err = git.GetCommitAtTag(tag)
	if err != nil {
		return tag, "", err
	}
	commitLatest, err = git.GetLatestCommit()
	if err != nil {
		return tag, commitAtTag, err
	}
	if commitLatest != commitAtTag {
		slog.Warn("This repo changed after application of", "tag", tag)
		slog.Warn("               at that tag", "commit", commitAtTag)
		slog.Warn("     does not match latest", "commit", commitLatest)
		slog.Warn("You probably want to apply a new tag.")
		err = fmt.Errorf("commit mismatch")
	}
	return tag, commitLatest, err
}

func buildBinaries(dirSrc, dirOut, tag, commitHash string) error {
	goRunner := internal.NewGoRunner(dirSrc, dirOut, tag, commitHash)
	var n string
	var err error
	n, err = goRunner.Build(internal.OsLinux, internal.ArchAmd64)
	if err != nil {
		return err
	}
	fmt.Println("created " + n)
	return nil
}
