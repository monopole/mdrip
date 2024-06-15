package loader_test

import (
	"testing"

	. "github.com/monopole/mdrip/v2/internal/loader"
	"github.com/stretchr/testify/assert"
)

func TestMyFolderNaming(t *testing.T) {
	f1 := NewEmptyFile("f1")
	f2 := NewEmptyFile("f2")
	d1 := NewFolder("d1").AddFile(f1).AddFile(f2)
	assert.Equal(t, "d1", d1.Name())
	assert.Equal(t, FilePath("d1"), d1.Path())
	assert.Equal(t, "f1", f1.Name())
	assert.Equal(t, FilePath("d1/f1"), f1.Path())
	assert.Equal(t, "f2", f2.Name())
	assert.Equal(t, FilePath("d1/f2"), f2.Path())

	d2 := NewFolder("d2").AddFolder(d1)
	assert.Equal(t, "d2", d2.Name())
	assert.Equal(t, FilePath("d2"), d2.Path())
	assert.Equal(t, "f2", f2.Name())
	assert.Equal(t, FilePath("d2/d1/f2"), f2.Path())
}

func TestMyFolderEquals(t *testing.T) {
	f1, f1Prime := NewEmptyFile("f1"), NewEmptyFile("f1")
	f2, f2Prime := NewEmptyFile("f2"), NewEmptyFile("f2")

	d1 := NewFolder("d1").AddFile(f1).AddFile(f2)
	d1Prime := NewFolder("d1").AddFile(f1Prime).AddFile(f2Prime)

	assert.True(t, d1.Equals(d1))
	assert.True(t, d1.Equals(d1Prime))

	d2 := NewFolder("d2").AddFolder(d1)
	d2Prime := NewFolder("d2").AddFolder(d1Prime)

	assert.True(t, d2.Equals(d2))
	assert.True(t, d2.Equals(d2Prime))
	assert.False(t, d2.Equals(d1))
}
