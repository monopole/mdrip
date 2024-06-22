package navleftroot_test

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"testing"

	"github.com/monopole/mdrip/v2/internal/webapp/widget/testutil"
	"github.com/spf13/afero"

	"github.com/monopole/mdrip/v2/internal/loader"
	. "github.com/monopole/mdrip/v2/internal/webapp/widget/navleftroot"
	"github.com/stretchr/testify/assert"
	"github.com/yosssi/gohtml"
)

func TestRenderer(t *testing.T) {
	type testCase struct {
		input          loader.MyTreeNode
		maxFileNameLen int
		want           string
	}
	for n, tc := range map[string]testCase{
		"t0": {
			input:          loader.NewFolder("DIR_0"),
			maxFileNameLen: 0,
			want: (`
<div class='navLeftFolder' id='navLeftFolderId0'>
  <div class='navLeftFolderName'>
    DIR_0
  </div>
  <div class='navLeftFolderChildren'></div>
</div>`)[1:],
		},
		"t0.5": {
			input:          loader.NewEmptyFile("B234567"),
			maxFileNameLen: 7,
			want: (`
<div class='navLeftFile navLeftFileDeactivated' id='navLeftFileId0'>
  B234567
</div>`)[1:],
		},
		"t1": {
			input: loader.NewFolder("DIR_0").
				AddFile(loader.NewEmptyFile("FILE_0")).
				AddFile(loader.NewEmptyFile("FILE_1")),
			maxFileNameLen: 8, /* (depth * 2) + 6 */
			want: (`
<div class='navLeftFolder' id='navLeftFolderId0'>
  <div class='navLeftFolderName'>
    DIR_0
  </div>
  <div class='navLeftFolderChildren'>
    <div class='navLeftFile navLeftFileDeactivated' id='navLeftFileId0'>
      FILE_0
    </div>
    <div class='navLeftFile navLeftFileDeactivated' id='navLeftFileId1'>
      FILE_1
    </div>
  </div>
</div>`)[1:],
		},
		"t2": {
			input: loader.NewTopFolder(loader.NewFolder("DIR_0").
				AddFile(loader.NewEmptyFile("FILE_0")).
				AddFile(loader.NewEmptyFile("FILE_1"))),
			maxFileNameLen: 8,
			want: (`
<div class='navLeftTopFolder'>
  <div class='navLeftFile navLeftFileDeactivated' id='navLeftFileId0'>
    FILE_0
  </div>
  <div class='navLeftFile navLeftFileDeactivated' id='navLeftFileId1'>
    FILE_1
  </div>
</div>`)[1:],
		},
		"t3": {
			input:          testutil.MakeFolderTreeOfMarkdown(),
			maxFileNameLen: 14, /* (depth==8 * 2) + 6 */
			want: `
<div class='navLeftFolder' id='navLeftFolderId0'>
  <div class='navLeftFolderName'>
    top
  </div>
  <div class='navLeftFolderChildren'>
    <div class='navLeftFile navLeftFileDeactivated' id='navLeftFileId0'>
      file00
    </div>
    <div class='navLeftFile navLeftFileDeactivated' id='navLeftFileId1'>
      file01
    </div>
    <div class='navLeftFile navLeftFileDeactivated' id='navLeftFileId2'>
      file02
    </div>
    <div class='navLeftFolder' id='navLeftFolderId1'>
      <div class='navLeftFolderName'>
        dir0
      </div>
      <div class='navLeftFolderChildren'>
        <div class='navLeftFile navLeftFileDeactivated' id='navLeftFileId3'>
          file03
        </div>
        <div class='navLeftFile navLeftFileDeactivated' id='navLeftFileId4'>
          file04
        </div>
        <div class='navLeftFolder' id='navLeftFolderId2'>
          <div class='navLeftFolderName'>
            dir1
          </div>
          <div class='navLeftFolderChildren'>
            <div class='navLeftFile navLeftFileDeactivated' id='navLeftFileId5'>
              file05
            </div>
            <div class='navLeftFile navLeftFileDeactivated' id='navLeftFileId6'>
              file06
            </div>
          </div>
        </div>
      </div>
    </div>
    <div class='navLeftFolder' id='navLeftFolderId3'>
      <div class='navLeftFolderName'>
        dir2
      </div>
      <div class='navLeftFolderChildren'>
        <div class='navLeftFile navLeftFileDeactivated' id='navLeftFileId7'>
          file07
        </div>
        <div class='navLeftFile navLeftFileDeactivated' id='navLeftFileId8'>
          file08
        </div>
        <div class='navLeftFile navLeftFileDeactivated' id='navLeftFileId9'>
          file09
        </div>
        <div class='navLeftFolder' id='navLeftFolderId4'>
          <div class='navLeftFolderName'>
            dir3
          </div>
          <div class='navLeftFolderChildren'>
            <div class='navLeftFile navLeftFileDeactivated' id='navLeftFileId10'>
              file10
            </div>
            <div class='navLeftFile navLeftFileDeactivated' id='navLeftFileId11'>
              file11
            </div>
            <div class='navLeftFile navLeftFileDeactivated' id='navLeftFileId12'>
              file12
            </div>
            <div class='navLeftFolder' id='navLeftFolderId5'>
              <div class='navLeftFolderName'>
                dir4
              </div>
              <div class='navLeftFolderChildren'>
                <div class='navLeftFile navLeftFileDeactivated' id='navLeftFileId13'>
                  file13
                </div>
                <div class='navLeftFile navLeftFileDeactivated' id='navLeftFileId14'>
                  file14
                </div>
                <div class='navLeftFile navLeftFileDeactivated' id='navLeftFileId15'>
                  file15
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
    <div class='navLeftFolder' id='navLeftFolderId6'>
      <div class='navLeftFolderName'>
        dir5
      </div>
      <div class='navLeftFolderChildren'>
        <div class='navLeftFile navLeftFileDeactivated' id='navLeftFileId16'>
          file16
        </div>
      </div>
    </div>
  </div>
</div>`[1:],
		},
	} {
		t.Run(n, func(t *testing.T) {
			var b bytes.Buffer
			r := NewRenderer(&b)
			tc.input.Accept(r)
			got := gohtml.Format(b.String()) // Make tests easier to read.
			assert.Equal(t, tc.maxFileNameLen, r.MaxFileNameLength())
			if !assert.Equal(t, tc.want, got) {
				fmt.Fprintln(os.Stderr, "--------------------")
				fmt.Fprintln(os.Stderr, got)
				fmt.Fprintln(os.Stderr, "--------------------")
			}
		})
	}
}

const runTheUnportableLocalFileSystemDependentTests = false

func TestFromDiskRenderer(t *testing.T) {
	if !runTheUnportableLocalFileSystemDependentTests {
		t.Skip("skipping non-portable tests")
	}
	f, err := loader.New(afero.NewOsFs()).LoadOneTree(
		"/home/yadayada/myrepos/github.com/sigs.k8s.io/kustomize")
	if err != nil {
		t.Fatal(err)
	}
	var b bytes.Buffer
	f.Accept(NewRenderer(&b))
	got := gohtml.Format(b.String()) // Make tests easier to read.
	_, _ = fmt.Fprintln(os.Stderr, got)
}
