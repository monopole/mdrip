package tutorial

import (
	"testing"

	"github.com/monopole/mdrip/model"
)

// TODO: add some real tests
func TestReload(t *testing.T) {
	NewProgramFromPaths(model.Label("foo"), []model.FilePath{})
}
