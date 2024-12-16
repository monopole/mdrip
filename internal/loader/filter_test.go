package loader_test

import (
	"io/fs"
	"os"
	"testing"
	"time"

	. "github.com/monopole/mdrip/v2/internal/loader"
	"github.com/stretchr/testify/assert"
)

type mockFileInfo struct {
	name string
	mode fs.FileMode
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return 0 }
func (m *mockFileInfo) Mode() fs.FileMode  { return m.mode }
func (m *mockFileInfo) ModTime() time.Time { panic("didn't think ModTime was needed") }
func (m *mockFileInfo) IsDir() bool        { panic("didn't think IsDir was needed") }
func (m *mockFileInfo) Sys() any           { panic("didn't think Sys was needed") }

var _ os.FileInfo = &mockFileInfo{}

type tCase struct {
	fi  *mockFileInfo
	err error
}

func TestIsMarkDownFile(t *testing.T) {
	for n, tc := range map[string]tCase{
		"t1": {
			fi: &mockFileInfo{
				name: "aDirectory.md",
				mode: fs.ModeDir,
			},
			err: NotMarkDownErr,
		},
		"t2": {
			fi: &mockFileInfo{
				name: "notMarkdown",
			},
			err: NotMarkDownErr,
		},
		"t3": {
			fi: &mockFileInfo{
				name: "aFileButIrregular.md",
				mode: fs.ModeIrregular,
			},
			err: NotMarkDownErr,
		},
		"t4": {
			fi: &mockFileInfo{
				name: "aFile.md",
			},
		},
	} {
		t.Run(n, func(t *testing.T) {
			assert.Equal(t, tc.err, IsMarkDownFile(tc.fi))
		})
	}
}

func TestInNotIgnorableFolder(t *testing.T) {
	for n, tc := range map[string]tCase{
		"t1": {
			fi: &mockFileInfo{
				name: ".git",
			},
			err: IsADotDirErr,
		},
		"t2": {
			fi: &mockFileInfo{
				name: "./",
			},
		},
		"t3": {
			fi: &mockFileInfo{
				name: "..",
			},
		},
		"iHateNode": {
			fi: &mockFileInfo{
				name: "node_modules",
			},
			err: IsANodeCache,
		},
	} {
		t.Run(n, func(t *testing.T) {
			assert.Equal(t, tc.err, InNotIgnorableFolder(tc.fi))
		})
	}
}
