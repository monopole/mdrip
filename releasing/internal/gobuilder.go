package internal

import (
	"os"
	"path/filepath"
	"strings"
	"time"
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

// GoBuilder runs `go build` and compresses/tars the result.
type GoBuilder struct {
	goRun   *MyRunner
	zipRun  *MyRunner
	tarRun  *MyRunner
	pgmName string
	dirOut  string
	ldVars  *LdVars
}

func NewGoBuilder(dirSrc, dirOut string, ldVars *LdVars) *GoBuilder {
	return &GoBuilder{
		goRun:   NewMyRunner("go", dirSrc, DoIt, 30*time.Second),
		zipRun:  NewMyRunner("zip", dirOut, DoIt, 30*time.Second),
		tarRun:  NewMyRunner("tar", dirOut, DoIt, 30*time.Second),
		pgmName: filepath.Base(dirSrc),
		dirOut:  dirOut,
		ldVars:  ldVars,
	}
}

func (gb *GoBuilder) Build(myOs EnumOs, myArch EnumArch) (string, error) {
	name := gb.binaryName(myOs)
	gb.goRun.comment(
		"building " + name + " for " + myOs.String() + ":" + myArch.String())
	gb.goRun.setEnv(map[string]string{
		"HOME":        os.Getenv("HOME"), // Make it easier to find ~/go/pkg
		"CGO_ENABLED": "0",               // Force static binaries.
		"GOOS":        myOs.String(),
		"GOARCH":      myArch.String(),
	})
	binary := filepath.Join(gb.dirOut, name)
	if err := gb.goRun.run(NoHarmDone,
		"build",
		"-o", binary,
		"-ldflags", gb.ldVars.makeLdFlags(),
		".", // Using the "." is why we need HOME defined.
	); err != nil {
		return name, err
	}
	p, err := gb.packageIt(myOs, myArch, name)
	_ = os.Remove(binary)
	return filepath.Join(gb.dirOut, p), err
}

func (gb *GoBuilder) binaryName(myOs EnumOs) string {
	if myOs == OsWindows {
		return gb.pgmName + ".exe"
	}
	return gb.pgmName
}

func (gb *GoBuilder) packageIt(
	myOs EnumOs, myArch EnumArch, fileName string) (string, error) {
	base := strings.Join([]string{
		gb.pgmName,
		gb.ldVars.version(),
		myOs.String(),
		myArch.String(),
	}, "_")
	if myOs == OsWindows {
		result := base + ".zip"
		return result, gb.zipRun.run(NoHarmDone,
			"-j", // suppress storing the full path to the file
			result,
			fileName,
		)
	}
	result := base + ".tar.gz"
	return result, gb.tarRun.run(NoHarmDone,
		"cfz", // create, verbose, fileName follows, compress with gzip
		result,
		fileName,
	)
}

// LdVars manages the value sent into `go build -ldflags {args}`.
// To see explanation, enter this crazy command:  go build -ldflags="-help"
type LdVars struct {
	ImportPath string
	Kvs        map[string]string
}

func (ldv *LdVars) makeDefinitions() []string {
	result := make([]string, len(ldv.Kvs))
	i := 0
	for k, v := range ldv.Kvs {
		result[i] = ldv.ImportPath + "." + k + "=" + v
		i++
	}
	return result
}

func (ldv *LdVars) makeLdFlags() string {
	result := []string{
		"-s", // disable symbol table (small binary)
		"-w", // disable DWARF generation (ditto)
	}
	defs := ldv.makeDefinitions()
	for i := range defs {
		result = append(result, "-X", defs[i])
	}
	return strings.Join(result, " ")
}

func (ldv *LdVars) version() string {
	v, ok := ldv.Kvs["version"]
	if !ok {
		panic("version not in ldFlags!")
	}
	return v
}
