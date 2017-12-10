package webapp

import (
	"bytes"
	"github.com/monopole/mdrip/base"
	"strings"
	"testing"
)

type waTest struct {
	name string
	want []string
}

var orderedPageParts = []string{
	"<head>",
	"<style",
	"header,",
	"navLeftBox",
	"helpButtonBox",
	"navBurger",
	"</style>",
	"<script",
	"getElByClass(",
	"navController =",
	"helpController =",
	"/script>",
	"/head>",
	"<body",
	"<header",
	"<div class='navLeftBox",
	"<div class='helpBox'",
	"<div class='scrollingColumn'",
	"<div class='proseColumn'",
	"</body>",
}

var waTests = []waTest{
	{"emptyTutorial", orderedPageParts},
}

func TestWebAppBasicTemplateRendered(t *testing.T) {
	ds, _ := base.NewDataSource("/tmp")
	wa := NewWebApp("", "", emptyLesson, ds, []int{}, [][]int{{}})
	for _, test := range waTests {

		var b bytes.Buffer
		wa.Render(&b)
		got := b.String()
		prev := 0
		for _, target := range test.want {
			got = got[prev:]
			k := strings.Index(got, target)
			prev = k
			if k < 0 {
				t.Errorf("%s:  Didn't find %s in page content\n-----\n%s\n-----",
					test.name, target, got)
				break
			}
		}
	}
}
