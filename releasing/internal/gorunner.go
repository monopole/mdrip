package internal

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	ldVarPath = "github.com/monopole/mdrip/internal/provenance"
)

//go:generate stringer -type=EnumOs -linecomment
type EnumOs int

const (
	OsUnknown EnumOs = iota // unknown
	OsLinux                 // linux
	OsDarwin                // darwin
	OsWindows               // windows
)

//go:generate stringer -type=EnumArch -linecomment
type EnumArch int

const (
	ArchUnknown EnumArch = iota // unknown
	ArchAmd64                   // amd64
	ArchArm64                   // arm64
)

// GoRunner runs some go commands.
type GoRunner struct {
	rn         *runner
	pgmName    string
	dirOut     string
	tag        string
	commitHash string
	timeStamp  time.Time
}

func NewGoRunner(dirSrc, dirOut, tag, commitHash string) *GoRunner {
	return &GoRunner{
		rn:         newRunner("go", dirSrc, DoIt, 30*time.Second),
		pgmName:    filepath.Base(dirSrc),
		dirOut:     dirOut,
		tag:        tag,
		commitHash: commitHash,
		timeStamp:  time.Now(),
	}
}

func (gr *GoRunner) Build(myOs EnumOs, myArch EnumArch) (string, error) {
	gr.rn.comment("building for " + myOs.String() + " " + myArch.String())

	binaryName := gr.pgmName
	if myOs == OsWindows {
		binaryName += ".exe"
	}
	result := filepath.Join(gr.dirOut, binaryName)
	gr.rn.setEnv(map[string]string{
		"HOME":        os.Getenv("HOME"),
		"CGO_ENABLED": "0",
		"GOOS":        myOs.String(),
		"GOARCH":      myArch.String(),
	})
	return result, gr.rn.run(noHarmDone,
		"build",
		"-o", result,
		"-ldflags", gr.generateLdFlags(),
		".",
	)
}

func (gr *GoRunner) generateLdFlags() string {
	result := []string{
		"-s",
		"-w",
		"-X", ldVarPath + ".version=" + gr.tag,
		"-X", ldVarPath + ".gitCommit=" + gr.commitHash,
		"-X", ldVarPath + ".buildDate=" + gr.timeStamp.Format(time.RFC3339),
	}
	return strings.Join(result, " ")
}
