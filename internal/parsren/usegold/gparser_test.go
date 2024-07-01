package usegold_test

import (
	_ "embed"
	"fmt"
	"testing"

	"github.com/monopole/mdrip/v2/internal/loader"
	"github.com/monopole/mdrip/v2/internal/parsren"
	. "github.com/monopole/mdrip/v2/internal/parsren/usegold"
	"github.com/monopole/mdrip/v2/internal/web/app/widget/codeblock"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

const (
	blockIdx = 3

	// noLang means no language specified in code block.
	noLang = ""
	// Authoring markdown in Go constants isn't fun; will use files for
	// bigger tests.

	smallMdExampleName    = "smallEx"
	smallMdExampleContent = `
# header

[google]: https://www.google.com

Some text before a code block.

<!-- @theOne  @two  @three -->
` + "```" + `
echo alpha
which find
` + "```" + `

An indented code block should be not be recognized as a code block.

> ` + "```" + `
> echo beta
> which ls
> ` + "```" + `

A comment between the code blocks.

<!-- @myFour @leFive -->
` + "```" + `
echo beta
which ls
` + "```" + `

The next block has no labels.

` + "```" + `
echo gamma
which cat
` + "```" + `

The end.
`
	tinyMdExampleName  = "tinyEx"
	tinyExampleContent = `
# header
Some text before a code block.
` + "```" + `
echo alpha
which find
` + "```" + `
The end.
`
)

const prompt = string(codeblock.CbPrompt)

func TestRenderingHtmlFromStringConstants(t *testing.T) {
	tests := map[string]struct {
		file *loader.MyFile
		path loader.FilePath
		html string
	}{
		"empty": {
			file: loader.NewEmptyFile("peach"),
			path: "peach",
			html: "",
		},
		"tinyNoLabels": {
			file: loader.NewFile(tinyMdExampleName, []byte(tinyExampleContent)),
			path: tinyMdExampleName,
			html: (`
<h1 id="header">header</h1>
<p>Some text before a code block.</p>
<div class='codeBlockContainer' id='codeBlockId0'>
  <div class='codeBlockControl'>
    <span class='codeBlockTitle'> codeBlock000 </span>
  </div>
  <div class='codeBlockPrompt'> ` + prompt + ` </div>
  <div class='codeBlockArea'>
echo alpha
which find
</div>
</div>
<p>The end.</p>
`)[1:],
		},
		"smallWithLabels": {
			file: loader.NewFile(
				smallMdExampleName, []byte(smallMdExampleContent)),
			path: smallMdExampleName,
			html: (`
<h1 id="header">header</h1>
<p>Some text before a code block.</p>
<!-- @theOne  @two  @three -->
<div class='codeBlockContainer' id='codeBlockId0'>
  <div class='codeBlockControl'>
    <span class='codeBlockTitle'> theOne </span>
  </div>
  <div class='codeBlockPrompt'> ` + prompt + ` </div>
  <div class='codeBlockArea'>
echo alpha
which find
</div>
</div>
<p>An indented code block should be not be recognized as a code block.</p>
<blockquote>
<pre><code>echo beta
which ls
</code></pre>
</blockquote>
<p>A comment between the code blocks.</p>
<!-- @myFour @leFive -->
<div class='codeBlockContainer' id='codeBlockId1'>
  <div class='codeBlockControl'>
    <span class='codeBlockTitle'> myFour </span>
  </div>
  <div class='codeBlockPrompt'> ` + prompt + ` </div>
  <div class='codeBlockArea'>
echo beta
which ls
</div>
</div>
<p>The next block has no labels.</p>
<div class='codeBlockContainer' id='codeBlockId2'>
  <div class='codeBlockControl'>
    <span class='codeBlockTitle'> codeBlock002 </span>
  </div>
  <div class='codeBlockPrompt'> ` + prompt + ` </div>
  <div class='codeBlockArea'>
echo gamma
which cat
</div>
</div>
<p>The end.</p>
`)[1:],
		},
	}
	for n, tc := range tests {
		t.Run(n, func(t *testing.T) {
			p := NewGParser()
			tc.file.Accept(p)
			assert.Equal(t, 1, len(p.RenderedMdFiles()))
			file := p.RenderedMdFiles()[0]
			assert.Equal(t, tc.path, file.Path)
			assert.Equal(t, tc.html, string(file.Html))
		})
	}
}

func TestParsingBlocksFromStringConstants(t *testing.T) {
	tests := map[string]struct {
		file           *loader.MyFile
		totBlocks      int
		label          loader.Label
		filteredBlocks []*loader.CodeBlock
	}{
		"empty": {
			file: loader.NewEmptyFile("peach"),
		},
		"one": {
			file:      loader.NewFile(smallMdExampleName, []byte(smallMdExampleContent)),
			label:     "theOne",
			totBlocks: 3,
			filteredBlocks: []*loader.CodeBlock{
				loader.NewCodeBlock(nil, `
echo alpha
which find
`[1:], blockIdx, noLang, "theOne", "two", "three"),
			},
		},
		"five": {
			file:      loader.NewFile(smallMdExampleName, []byte(smallMdExampleContent)),
			label:     "leFive",
			totBlocks: 3,
			filteredBlocks: []*loader.CodeBlock{
				loader.NewCodeBlock(nil, `
echo beta
which ls
`[1:], blockIdx, noLang, "myFour", "leFive"),
			},
		},
		"nope": {
			file:           loader.NewFile(smallMdExampleName, []byte(smallMdExampleContent)),
			totBlocks:      3,
			label:          "nope",
			filteredBlocks: []*loader.CodeBlock{},
		},
	}
	for n, tc := range tests {
		t.Run(n, func(t *testing.T) {
			p := NewGParser()
			tc.file.Accept(p)
			if !assert.Equal(t, 1, len(p.RenderedMdFiles())) {
				t.FailNow()
			}
			blocks := p.RenderedMdFiles()[0].Blocks
			assert.Equal(t, tc.totBlocks, len(blocks))
			blocks = p.FilteredBlocks(tc.label)
			if !assert.Equal(t, len(tc.filteredBlocks), len(blocks)) {
				t.FailNow()
			}
			for i := range blocks {
				assert.True(t, tc.filteredBlocks[i].Equals(blocks[i]))
			}
		})
	}
}

func TestParsingTree(t *testing.T) {
	var p parsren.MdParserRenderer
	{ // TODO: try embedding the file system
		folder, err := loader.New(afero.NewOsFs()).LoadFolder("testdata")
		assert.NoError(t, err)
		p = NewGParser()
		folder.Accept(p)
	}
	if !assert.Equal(t, 2, len(p.RenderedMdFiles())) {
		t.FailNow()
	}
	if !assert.Equal(t, 6, len(p.FilteredBlocks(loader.WildCardLabel))) {
		t.FailNow()
	}
	if printTheHtml := false; printTheHtml {
		fmt.Println("<html><body>")
		for _, f := range p.RenderedMdFiles() {
			fmt.Println(f.Html)
			fmt.Println("<!-- ------------- -->")
		}
		fmt.Println("</body></html>")
	}
}
