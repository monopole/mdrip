package loader_test

import (
	"fmt"
	"log/slog"
	"os"
	"testing"

	. "github.com/monopole/mdrip/v2/internal/loader"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

var debugging = false

func turnOnDebugging() {
	debugging = true
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey || a.Key == slog.LevelKey {
				a.Value = slog.StringValue("")
			}
			return a
		},
	})))
}

// Permission bits
//
//	The file or folder's owner:
//
// 400  Read
// 200  Write
// 100  Execute/search
//
//	    Other users in the file or folder's group:
//	40  Read
//	20  Write
//	10  Execute/search
//
//	    Other users not in the group:
//	 4  Read
//	 2  Write
//	 1  Execute/search
const (
	RW  os.FileMode = 0644
	RWX os.FileMode = 0755
)

var (
	md       []*MyFile
	readmeMd = NewFile(ReadmeFileName, []byte("# Howdy!"))
)

// Define a bunch of markdown files and their contents.
// A file whose name ends in ".md" is considered a markdown file.
func init() {
	md = make([]*MyFile, 11)
	for i := range md {
		md[i] = NewFile(fmt.Sprintf("f%02d.md", i), []byte(fmt.Sprintf("# file f%02d", i)))
	}
}

func makeSmallAbsFs(t *testing.T, fs afero.Fs) {
	assert.NoError(t, afero.WriteFile(fs, "/f00.md", md[0].C(), RW))
	assert.NoError(t, afero.WriteFile(fs, "/aaa/f01.md", md[1].C(), RW))
}

func makeSmallRelFs(t *testing.T, fs afero.Fs) {
	assert.NoError(t, afero.WriteFile(fs, "f00.md", md[0].C(), RW))
	assert.NoError(t, afero.WriteFile(fs, "aaa/f01.md", md[1].C(), RW))
}

func makeMediumAbsFs(t *testing.T, fs afero.Fs) {
	assert.NoError(t, afero.WriteFile(fs, "/f00.md", md[0].C(), RW))
	assert.NoError(t, afero.WriteFile(fs, "/aaa/bbb/f01.md", md[1].C(), RW))
	assert.NoError(t, afero.WriteFile(fs, "/aaa/f02.md", md[2].C(), RW))
	assert.NoError(t, afero.WriteFile(fs, "/aaa/ccc/f03.md", md[3].C(), RW))
	assert.NoError(t, afero.WriteFile(fs, "/aaa/ccc/ignore", []byte("not markdown"), RW))
	assert.NoError(t, fs.MkdirAll("/aaa/empty", RWX))
}

func makeMediumRelFs(t *testing.T, fs afero.Fs) {
	assert.NoError(t, afero.WriteFile(fs, "f00.md", md[0].C(), RW))
	assert.NoError(t, afero.WriteFile(fs, "aaa/bbb/f01.md", md[1].C(), RW))
	assert.NoError(t, afero.WriteFile(fs, "aaa/f02.md", md[2].C(), RW))
	assert.NoError(t, afero.WriteFile(fs, "aaa/ccc/f03.md", md[3].C(), RW))
	assert.NoError(t, afero.WriteFile(fs, "aaa/ccc/ignore", []byte("not markdown"), RW))
	assert.NoError(t, fs.MkdirAll("aaa/empty", RWX))
}

func makeLargeAbsFs(t *testing.T, fs afero.Fs) {
	assert.NoError(t, afero.WriteFile(fs, "/f10.md", md[10].C(), RW))
	assert.NoError(t, afero.WriteFile(fs, "/mmm/yyy/f09.md", md[9].C(), RW))
	assert.NoError(t, afero.WriteFile(fs, "/mmm/f08.md", md[8].C(), RW))
	assert.NoError(t, afero.WriteFile(fs, "/mmm/eee/f07.md", md[7].C(), RW))
	assert.NoError(t, afero.WriteFile(fs, "/mmm/eee/f06.md", md[6].C(), RW))
	assert.NoError(t, afero.WriteFile(fs, "/mmm/eee/ignore", []byte("not markdown"), RW))
	assert.NoError(t, fs.MkdirAll("/mmm/empty", RWX))
	assert.NoError(t, afero.WriteFile(fs, "/jjj/bbb/f05.md", md[5].C(), RW))
	assert.NoError(t, afero.WriteFile(fs, "/jjj/f04.md", md[4].C(), RW))
	assert.NoError(t, afero.WriteFile(fs, "/jjj/ccc/f03.md", md[3].C(), RW))
	assert.NoError(t, afero.WriteFile(fs, "/jjj/ccc/f02.md", md[2].C(), RW))
	assert.NoError(t, afero.WriteFile(fs, "/jjj/aaa/f01.md", md[1].C(), RW))
	assert.NoError(t, afero.WriteFile(fs, "/jjj/aaa/f00.md", md[0].C(), RW))
}

func TestAferoNonRootPath(t *testing.T) {
	// There's no notion of a "current" working folder
	// that you can change when writing to afero.
	// You can read and write absolute paths, which always start with RootSlash,
	// or you can read/write paths that don't start with RootSlash.
	// There's no way to "cd" in the RootSlash file system,
	// and have that change the behavior read/write.
	fs := afero.NewMemMapFs() // afero.NewOsFs()
	assert.NoError(t, afero.WriteFile(fs, "f00.md", md[0].C(), RW))
	assert.NoError(t, afero.WriteFile(fs, "/f01.md", md[1].C(), RW))
	data, err := afero.ReadFile(fs, "f00.md")
	assert.NoError(t, err)
	assert.Equal(t, md[0].C(), data)
	data, err = afero.ReadFile(fs, "/f01.md")
	assert.NoError(t, err)
	assert.Equal(t, md[1].C(), data)
	_, err = afero.ReadFile(fs, "f01.md")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file does not exist")
}

func TestLoadFolderFromMemoryHappy(t *testing.T) {
	type testC struct {
		fillFs      func(*testing.T, afero.Fs)
		pathToLoad  FilePath
		expectedFld func() *MyFolder
		errMsg      string
	}
	for n, tc := range map[string]testC{
		"nothingOk": {
			fillFs: func(tt *testing.T, fs afero.Fs) {
				// don't make any files.
			},
			pathToLoad: "/",
			expectedFld: func() *MyFolder {
				return nil
			},
		},
		"nothingWithError": {
			fillFs: func(tt *testing.T, fs afero.Fs) {
				// don't make any files.
			},
			pathToLoad: "/a.md",
			errMsg:     "file does not exist",
		},
		"nonExistentFile": {
			fillFs:     makeSmallAbsFs,
			pathToLoad: "/monkey",
			errMsg:     "does not exist",
		},
		"noGoingUp": {
			fillFs:     makeSmallAbsFs,
			pathToLoad: "../zzz",
			errMsg:     "specify absolute path or something at or below your working folder",
		},
		"oneFile": {
			fillFs: func(tt *testing.T, fs afero.Fs) {
				assert.NoError(tt, afero.WriteFile(fs, "/f01.md", md[1].C(), RW))
			},
			pathToLoad: "/f01.md",
			expectedFld: func() *MyFolder {
				return NewFolder("/").AddFile(md[1])
			},
		},
		"oneFileButAskForWrongFile": {
			fillFs: func(tt *testing.T, fs afero.Fs) {
				assert.NoError(tt, afero.WriteFile(fs, "/f01.md", md[1].C(), RW))
			},
			pathToLoad: "/f02.md",
			errMsg:     "file does not exist",
		},
		"oneEmptyFolder": {
			fillFs: func(tt *testing.T, fs afero.Fs) {
				assert.NoError(t, fs.MkdirAll("/aaa", RWX))
			},
			pathToLoad: "/",
			expectedFld: func() *MyFolder {
				return nil
			},
		},
		"oneEmptyFolderAgain": {
			fillFs: func(tt *testing.T, fs afero.Fs) {
				assert.NoError(t, fs.MkdirAll("/aaa", RWX))
			},
			pathToLoad: "/aaa",
			expectedFld: func() *MyFolder {
				return nil
			},
		},
		"justOneDirWithOneFile": {
			fillFs: func(tt *testing.T, fs afero.Fs) {
				assert.NoError(tt, afero.WriteFile(fs, "/aaa/f01.md", md[1].C(), RW))
			},
			pathToLoad: "/",
			expectedFld: func() *MyFolder {
				return NewFolder("/").AddFolder(NewFolder("aaa").AddFile(md[1]))
			},
		},
		"justOneSubDir": {
			fillFs: func(tt *testing.T, fs afero.Fs) {
				assert.NoError(tt, afero.WriteFile(fs, "/aaa/f01.md", md[1].C(), RW))
			},
			pathToLoad: "/aaa",
			expectedFld: func() *MyFolder {
				return NewFolder("/aaa").AddFile(md[1])
			},
		},
		"allOfSmallAbsFs": {
			fillFs:     makeSmallAbsFs,
			pathToLoad: "/",
			expectedFld: func() *MyFolder {
				aaa := NewFolder("aaa").AddFile(md[1])
				return NewFolder("/").AddFile(md[0]).AddFolder(aaa)
			},
		},
		"allOfSmallRelFs": {
			fillFs:     makeSmallRelFs,
			pathToLoad: ".",
			expectedFld: func() *MyFolder {
				aaa := NewFolder("aaa").AddFile(md[1])
				return NewFolder(".").AddFile(md[0]).AddFolder(aaa)
			},
		},
		"allOfSmallRelFsEmptyPath": {
			fillFs:     makeSmallRelFs,
			pathToLoad: "",
			expectedFld: func() *MyFolder {
				aaa := NewFolder("aaa").AddFile(md[1])
				return NewFolder(".").AddFile(md[0]).AddFolder(aaa)
			},
		},
		"fromMediumAbsEverything": {
			fillFs:     makeMediumAbsFs,
			pathToLoad: "/",
			expectedFld: func() *MyFolder {
				ccc := NewFolder("ccc").AddFile(md[3])
				bbb := NewFolder("bbb").AddFile(md[1])
				aaa := NewFolder("aaa").AddFile(md[2]).
					AddFolder(bbb).AddFolder(ccc)
				return NewFolder("/").AddFile(md[0]).AddFolder(aaa)
			},
		},
		"fromMediumRelEverything": {
			fillFs:     makeMediumRelFs,
			pathToLoad: ".",
			expectedFld: func() *MyFolder {
				ccc := NewFolder("ccc").AddFile(md[3])
				bbb := NewFolder("bbb").AddFile(md[1])
				aaa := NewFolder("aaa").AddFile(md[2]).
					AddFolder(bbb).AddFolder(ccc)
				return NewFolder(".").AddFile(md[0]).AddFolder(aaa)
			},
		},
		"fromMediumAbsJustAAA": {
			fillFs:     makeMediumAbsFs,
			pathToLoad: "/aaa",
			expectedFld: func() *MyFolder {
				ccc := NewFolder("ccc").AddFile(md[3])
				bbb := NewFolder("bbb").AddFile(md[1])
				return NewFolder("/aaa").AddFile(md[2]).
					AddFolder(bbb).AddFolder(ccc)
			},
		},
		"fromMediumAbsJustM0": {
			fillFs:     makeMediumAbsFs,
			pathToLoad: "/f00.md",
			expectedFld: func() *MyFolder {
				return NewFolder("/").AddFile(md[0])
			},
		},
		"fromMediumAbsJustM3": {
			fillFs:     makeMediumAbsFs,
			pathToLoad: "/aaa/ccc/f03.md",
			expectedFld: func() *MyFolder {
				return NewFolder("/aaa/ccc").AddFile(md[3])
			},
		},
		"fromLargeAbsEverything": {
			fillFs:     makeLargeAbsFs,
			pathToLoad: "/",
			expectedFld: func() *MyFolder {
				yyy := NewFolder("yyy").AddFile(md[9])
				eee := NewFolder("eee").AddFile(md[6]).AddFile(md[7])
				bbb := NewFolder("bbb").AddFile(md[5])
				ccc := NewFolder("ccc").AddFile(md[2]).AddFile(md[3])
				aaa := NewFolder("aaa").AddFile(md[0]).AddFile(md[1])
				mmm := NewFolder("mmm").AddFile(md[8]).AddFolder(eee).AddFolder(yyy)
				jjj := NewFolder("jjj").AddFile(md[4]).AddFolder(aaa).AddFolder(bbb).AddFolder(ccc)
				return NewFolder("/").AddFile(md[10]).AddFolder(jjj).AddFolder(mmm)
			},
		},
		"fromLargeAbsEverythingWithReordering": {
			fillFs: func(tt *testing.T, fs afero.Fs) {
				makeLargeAbsFs(tt, fs)
				assert.NoError(tt, afero.WriteFile(fs, "/jjj/"+OrderingFileName, []byte("ccc\nbbb\naaa"), RW))
				assert.NoError(tt, afero.WriteFile(fs, "/mmm/eee/"+readmeMd.Name(), readmeMd.C(), RW))
			},
			pathToLoad: "/",
			expectedFld: func() *MyFolder {
				yyy := NewFolder("yyy").AddFile(md[9])
				eee := NewFolder("eee").AddFile(readmeMd).AddFile(md[6]).AddFile(md[7])
				bbb := NewFolder("bbb").AddFile(md[5])
				ccc := NewFolder("ccc").AddFile(md[2]).AddFile(md[3])
				aaa := NewFolder("aaa").AddFile(md[0]).AddFile(md[1])
				mmm := NewFolder("mmm").AddFile(md[8]).AddFolder(eee).AddFolder(yyy)
				jjj := NewFolder("jjj").AddFile(md[4]).AddFolder(ccc).AddFolder(bbb).AddFolder(aaa)
				return NewFolder("/").AddFile(md[10]).AddFolder(jjj).AddFolder(mmm)
			},
		},
	} {
		t.Run(n, func(t *testing.T) {
			fs := afero.NewMemMapFs() // afero.NewOsFs()
			tc.fillFs(t, fs)
			ldr := New(fs)
			fld, err := ldr.LoadFolder(tc.pathToLoad)
			if tc.errMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errMsg)
				return
			}
			assert.NoError(t, err)
			if !assert.True(t, tc.expectedFld().Equals(fld)) {
				t.Errorf("Didn't get expected folder.")
				t.Log("Loaded:")
				fld.Accept(NewVisitorDump(os.Stdout))
				t.Log("Expected:")
				tc.expectedFld().Accept(NewVisitorDump(os.Stdout))
			}
		})
	}
}

const runTheUnportableLocalFileSystemDependentTests = false

func TestLoadOneTree(t *testing.T) {
	if !runTheUnportableLocalFileSystemDependentTests {
		t.Skip("skipping non-portable tests")
	}
	repoRoot := "/home/" + os.Getenv("USER") + "/myrepos/github.com"
	type testC struct {
		arg     FilePath
		topName string
		errMsg  string
	}
	for n, tc := range map[string]testC{
		"t1": {
			arg:     "README.md",
			topName: ".",
		},
		"t3": {
			arg:     FilePath(repoRoot + "/monopole/mdrip/testdata/file00.md"),
			topName: repoRoot + "/monopole/mdrip/testdata",
		},
		"t4": {
			arg:     ".",
			topName: ".",
		},
		"t5": {
			arg:    "/etc/passwd",
			errMsg: "not a simple markdown file",
		},
		"t6": {
			arg:    "/etc",
			errMsg: "unable to read folder",
		},
	} {
		t.Run(n, func(t *testing.T) {
			fsl := New(afero.NewOsFs())
			f, err := fsl.LoadOneTree(tc.arg)
			if tc.errMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errMsg)
				return
			}
			assert.NoError(t, err)
			if debugging {
				f.Accept(NewVisitorDump(os.Stdout))
			}
			assert.Equal(t, tc.topName, f.Name())
		})
	}
}

const runRealGitHubTests = false

func TestLoadOneTreeFromRepo(t *testing.T) {
	if !runRealGitHubTests {
		t.Skip("skipping real github tests")
	}
	turnOnDebugging()
	type testC struct {
		arg     FilePath
		topName string
	}
	for n, tc := range map[string]testC{
		"gh1": {
			arg:     "git@github.com:monopole/mdrip.git",
			topName: "git@github.com:monopole/mdrip",
		},
		"gh2": {
			arg:     "git@github.com:monopole/mdrip.git/README.md",
			topName: "git@github.com:monopole/mdrip",
		},
		"gh3": {
			arg:     "git@github.com:monopole/mdrip.git/data",
			topName: "git@github.com:monopole/mdrip/v2/data",
		},
		"gh4": {
			arg:     "git@github.com:monopole/mdrip.git/data/bad.md",
			topName: "git@github.com:monopole/mdrip/v2/data",
		},
	} {
		t.Run(n, func(t *testing.T) {
			// Must use a real FS, since the git command is used and it clones to real FS.
			fsl := New(afero.NewOsFs())
			f, err := fsl.LoadOneTree(tc.arg)
			assert.NoError(t, err)
			if debugging {
				f.Accept(NewVisitorDump(os.Stdout))
			}
			assert.Equal(t, tc.topName, f.Name())
		})
	}
}
