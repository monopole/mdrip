package program

import (
	"testing"

	"github.com/monopole/mdrip/model"
)

// TODO: add some real tests
func TestReload(t *testing.T) {
	p := NewProgram(model.Label("foo"), []model.FilePath{})
	p.Reload()
}
