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
	"div.titleBar",
	"div.leftNav",
	"div.lessonList",
	"div.oneLesson",
	"</style>",
	"<script",
	"toggleLeftNav(",
	"onLoad()",
	"/script>",
	"/head>",
	"<body",
	"<div class='instructions'",
	"<div class='titleBar'",
	"<div class='leftNav'",
	"<div class='lessonList'",
	"</body>",
}

var waTests = []waTest{
	{"emptyTutorial",
		NewWebApp("", "", emptyLesson, []int{}),
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
