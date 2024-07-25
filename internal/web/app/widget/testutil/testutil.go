package testutil

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monopole/mdrip/v2/internal/utils"

	"github.com/monopole/mdrip/v2/internal/loader"
	"github.com/monopole/mdrip/v2/internal/parsren/usegold"
	"github.com/monopole/mdrip/v2/internal/web/app/widget/appstate"
	"github.com/monopole/mdrip/v2/internal/web/app/widget/common"
	"github.com/stretchr/testify/assert"
)

type AppState struct {
	Stories []string
	Current int
}

var (
	//go:embed testutil.css
	Css string

	//go:embed testutil.js
	Js string
)

// For testing.
const (
	// Set true to save the renders to a file, set false to merely execute
	// templates and triggers any field errors.
	sendWidgetToFile = true

	// Set to true to write the file to the same directory as the test
	// (making it easy to render it in the IDE), set false to render test runs
	// into the same path under Documents.
	useLocalFile = false

	// File name to use when writing a widget's HTML for testing it.
	widgetFileName = "widget.html"

	// The name of the env var holding the directory into which these
	// tests should write widgetFileName when useLocalFile == false.
	envVarWidgetDir = "MDRIP_TEST_DIR"

	TmplTestName = "testableTemplate"
)

// fakeFile is for testing.
type fakeFile struct {
	bytes.Buffer
}

// Close is for testing.
func (f *fakeFile) Close() error {
	return nil
}

// getWriteCloser is for testing.
func getWriteCloser(t *testing.T) io.WriteCloser {
	if !sendWidgetToFile {
		return &fakeFile{}
	}
	var (
		f   *os.File
		err error
	)
	dir := directoryForWritingRenderedMarkdown(t)
	if _, err = os.Stat(dir); err != nil {
		t.FailNow()
	}
	p := filepath.Join(dir, widgetFileName)
	f, err = os.Create(p)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	t.Logf("Writing to %s", p)
	return f
}

func directoryForWritingRenderedMarkdown(t *testing.T) string {
	if useLocalFile {
		return "."
	}
	tmp := os.Getenv(envVarWidgetDir)
	if tmp == "" {
		t.Fatalf(
			"to use non-local html files, define env var %q", envVarWidgetDir)
	}
	stat, err := utils.PathStatus(tmp)
	if err != nil || stat != utils.PathIsAFolder {
		t.Fatalf(
			"value of env var %q doesn't resolve to a folder", envVarWidgetDir)
	}
	return tmp
}

func RenderHtmlToFile(t *testing.T, tmplBare string, values any) {
	tmplParsed, err := common.ParseAsHtmlTemplate(tmplBare)
	assert.NoError(t, err)
	f := getWriteCloser(t)
	assert.NoError(t, tmplParsed.ExecuteTemplate(f, TmplTestName, values))
	assert.NoError(t, f.Close())
}

func RenderTextToFile(t *testing.T, tmplBare string, values any) {
	tmplParsed, err := common.ParseAsTextTemplate(tmplBare)
	assert.NoError(t, err)
	f := getWriteCloser(t)
	assert.NoError(t, tmplParsed.ExecuteTemplate(f, TmplTestName, values))
	assert.NoError(t, f.Close())
}

func MakeAppStateTest0() *appstate.AppState {
	return MakeAppStateTest1(MakeFolderTreeOfMarkdown())
}

func MakeAppStateTest1(folder loader.MyTreeNode) *appstate.AppState {
	return makeAppStateTest2(
		"/my/folder/of/markdown",
		folder,
		"On Her Majesty's Secret Service")
}

func makeAppStateTest2(
	dSource string, folder loader.MyTreeNode, title string) *appstate.AppState {
	v := usegold.NewGParser()
	folder.Accept(v)
	return appstate.New(dSource, v.RenderedMdFiles(), title)
}

func MakeFolderTreeOfMarkdown() *loader.MyFolder {
	return MakeNamedFolderTreeOfMarkdown(loader.NewFolder("top"))
}

func MakeNamedFolderTreeOfMarkdown(top *loader.MyFolder) *loader.MyFolder {
	return top.
		AddFile(loader.NewFile("file00.md", mdBytes(0))).
		AddFile(loader.NewFile("file01.md", mdBytes(1))).
		AddFile(loader.NewFile("file02.md", mdBytes(2))).
		AddFolder(loader.NewFolder("dir0").
			AddFile(loader.NewFile("file03.md", mdBytes(3))).
			AddFile(loader.NewFile("file04.md", mdBytes(4))).
			AddFolder(loader.NewFolder("dir1").
				AddFile(loader.NewFile("file05.md", mdBytes(5))).
				AddFile(loader.NewFile("file06.md", mdBytes(6))))).
		AddFolder(loader.NewFolder("dir2").
			AddFile(loader.NewFile("file07.md", mdBytes(7))).
			AddFile(loader.NewFile("file08.md", mdBytes(8))).
			AddFile(loader.NewFile("file09.md", mdBytes(9))).
			AddFolder(loader.NewFolder("dir3").
				AddFile(loader.NewFile("file10.md", mdBytes(10))).
				AddFile(loader.NewFile("file11.md", mdBytes(11))).
				AddFile(loader.NewFile("file12.md", mdBytes(12))).
				AddFolder(loader.NewFolder("dir4").
					AddFile(loader.NewFile("file13.md", mdBytes(13))).
					AddFile(loader.NewFile("file14.md", mdBytes(14))).
					AddFile(loader.NewFile("file15.md", mdBytes(15)))))).
		AddFolder(loader.NewFolder("dir5").
			AddFile(loader.NewFile("file16.md", mdBytes(16))))
}

// mdBytes returns the bytes of a markdown document.
// The id arg is just a number that should appear in the doc.
func mdBytes(id int) []byte {
	myFmt := randomFormat()
	var buff bytes.Buffer
	buff.WriteString(fmt.Sprintf("# MD Doc %d", id))
	limit := 1 + rand.Intn(10)
	for i := 0; i < limit; i++ {
		buff.WriteString(fmt.Sprintf(myFmt, id))
		buff.WriteString(randomLoremIpsum())
		buff.WriteString("\n")
		if often() {
			// Add some labels.
			buff.WriteString("<!-- ")
			buff.WriteString(randomLabel())
			buff.WriteString(" @test -->\n")
		}
		buff.WriteString("```sh\n")
		buff.WriteString(randomCodeBlock())
		buff.WriteString("```\n")
		if !often() {
			buff.WriteString("\n```\n")
			buff.WriteString(randomCodeBlock())
			buff.WriteString("```\n")
		}
		if !often() {
			buff.WriteString("\n<!-- @mississippi -->\n")
			buff.WriteString("\n```\n")
			buff.WriteString(randomCodeBlock())
			buff.WriteString("```\n")
		}
		buff.WriteString(randomLoremIpsum())
	}
	return buff.Bytes()
}

func FillerDiv(s string) template.HTML {
	return template.HTML(`<div class="filler"> ` + s + ` </div>`)
}

// often is true 75% of the time.
func often() bool { return rand.Intn(4) != 0 }

func randomFormat() string {
	switch rand.Intn(3) {
	case 0:
		return mdFmt0
	case 1:
		return mdFmt1
	default:
		return mdFmt2
	}
}

func randomLabel() string {
	return fmt.Sprintf("@%s%03d", elements[rand.Intn(len(elements))], rand.Intn(999))
}

func randomLoremIpsum() string {
	return loremIpsum[rand.Intn(len(loremIpsum))][1:]
}

func randomCodeBlock() string { return codeBlocks[rand.Intn(len(codeBlocks))][1:] }

func LoremIpsum(n int) template.HTML {
	var s bytes.Buffer
	for i := 1; i < n; i++ {
		for _, line := range loremIpsum {
			s.WriteString("<p>")
			s.WriteString(line)
			s.WriteString("</p>")
		}
	}
	return template.HTML(s.String())
}

const (
	mdFmt0 = `
## Frank %d

### Fly me to the __moon__

Let me play among the _stars_
Let me see what spring is like
On _Jupiter_ and _Mars_
In other words, hold my hand
In other words, baby, kiss me

`

	mdFmt1 = `
## Fake Sun Tzu %d

Hence, when we are able to attack, we must look like we're scanning tic-toc;
when using our forces, we must appear to be drowsy.

| x | y | z | planet |
|---|---|---|---------|
| a | b | c | jupiter |
| d | e | f | mars |
| g | h | i | venus |
| j | k | l | earth |

When we are near, we must make the enemy believe that we've gone out for coffee;
when far away, we must make him believe we are under the bed.

`

	mdFmt2 = `
## Have Space Suit - Will Travel %d

Women and cats will do as they please;
men and dogs should relax and get used to the idea.

> There is no worse tyranny than to force a man to pay 
> for what he does not want merely because you think
> it would be good for him.

Always store beer in a dark place.

> ` + "```" + `
> let a = b**2 + c**2
> print math.sqrt(a)
> ` + "```" + `

This is the Unix philosophy: Write programs that do one thing and do it well.
Write programs to work together. Write programs to handle text streams,
because that is a universal interface.
`
)

var (
	codeBlocks = []string{`
which ls
cat /etc/hosts
`, `
which cat
date
`, `
ls /etc | wc -l
echo Doug McIlroy, can you summarize what\'s most important?
`, `
echo "Greetings, program!"
time
date
cat /etc/hosts | wc -c
cal
`,
	}

	loremIpsum = []string{`
Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor
incididunt ut labore et dolore magna aliqua.
`, `
Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut
aliquip ex ea commodo consequat.
`, `
Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu
fugiat nulla pariatur.
`, `
Excepteur sint occaecat cupidatat non proident,sunt in culpa qui officia
deserunt mollit anim id est laborum.
`,
	}

	elements = strings.Split(`
Hydrogen
Helium
Lithium
Beryllium
Boron
Carbon
Nitrogen
Oxygen
Fluorine
Neon
Sodium
Magnesium
Aluminium
Silicon
Phosphorus
Sulfur
Chlorine
Argon
Potassium
Calcium
Scandium
Titanium
Vanadium
Chromium
Manganese
Iron
Cobalt
Nickel
Copper
Zinc
Gallium
Germanium
Arsenic
Selenium
Bromine
Krypton
Rubidium
Strontium
Yttrium
Zirconium
Niobium
Molybdenum
Technetium
Ruthenium
Rhodium
Palladium
Silver
Cadmium
Indium
Tin
Antimony
Tellurium
Iodine
Xenon
Cesium
Barium
Lanthanum
Cerium
Praseodymium
Neodymium
Promethium
Samarium
Europium
Gadolinium
Terbium
Dysprosium
Holmium
Erbium
Thulium
Ytterbium
Lutetium
Hafnium
Tantalum
Tungsten
Rhenium
Osmium
Iridium
Platinum
Gold
Mercury
Thallium
Lead
Bismuth
Polonium
Astatine
Radon
Francium
Radium
Actinium
Thorium
Protactinium
Uranium
Neptunium
Plutonium
Americium
Curium
Berkelium
Californium
Einsteinium
Fermium
Mendelevium
Nobelium
Lawrencium
Rutherfordium
Dubnium
Seaborgium
Bohrium
Hassium
Meitnerium
Darmstadtium
Roentgenium
Copernicium
Nihonium
Flerovium
Moscovium
Livermorium
Tennessine
Oganesson`[1:], "\n")
)
