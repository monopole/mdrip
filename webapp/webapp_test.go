package webapp

import (
	"bytes"
	"strings"
	"testing"
)

type waTest struct {
	name  string
	input *WebApp
	want  []string
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
	"nav =",
	"help =",
	"/script>",
	"/head>",
	"<body",
	"<header",
	"<div class='navLeftBox'",
	"<div class='helpBox'",
	"<div class='scrollingColumn'",
	"<div class='proseColumn'",
	"</body>",
}

var waTests = []waTest{
	{"emptyTutorial",
		NewWebApp("", "", emptyLesson, []int{}, [][]int{{}}),
		orderedPageParts},
}

func TestWebAppBasicTemplateRendered(t *testing.T) {
	for _, test := range waTests {
		var b bytes.Buffer
		test.input.Render(&b)
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
