package program

import (
	"testing"

	"github.com/monopole/mdrip/base"
	"github.com/monopole/mdrip/model"
)

func TestBlockPgm(t *testing.T) {
	bParsed := model.NewBlockParsed(
		[]base.Label{}, base.MdProse([]byte("_foo_")), base.OpaqueCode("bar"))
	bTut := model.NewBlockTut(bParsed)
	bPgm := NewBlockPgmFromBlockTut(bTut)
	got := string(bPgm.HtmlProse())
	expected := "<p><em>foo</em></p>\n"
	if got != expected {
		t.Errorf(
			"Expected \"%s\",\n" +
				"     got \"%s\"", expected, got)
	}
	got = bPgm.Name()
	expected = model.AnonBlockName
	if got != expected {
		t.Errorf("name expected \"%s\",\n" +
			"          got \"%s\"", expected, got)
	}
	got = string(bPgm.Code())
	expected = "bar"
	if got != expected {
		t.Errorf("name expected \"%s\",\n" +
			"          got \"%s\"", expected, got)
	}
}
