package tutorial

import (
	"fmt"
	"github.com/monopole/mdrip/model"
	"github.com/russross/blackfriday"
	"html/template"
	"io"
)

// Tutorial UX Overview.
//
// Suppose it's a tutorial on Benelux.
//
// The first lesson is an overview of Benelux, with sibling (_not_ child)
// lessons covering Belgium, Netherlands, and Luxembourg.  These in turn
// could contain lessons on provinces, which could contain lessons on
// cities, etc.
//
// Associated content REST addresses look like
//
//     benelux.com/overview                  // Describes Benelux in general.
//     benelux.com/history                   // Benelux history, economy, etc.
//     benelux.com/economy
//     benelux.com/belgium/overview          // Describes Belgium in general.
//     benelux.com/belgium/tintin            // Dive into important details.
//     benelux.com/belgium/beer
//     benelux.com/belgium/antwerp/overview  // Dive into Antwerp, etc.
//     benelux.com/belgium/antwerp/diamonds
//     benelux.com/belgium/antwerp/rubens
//     benelux.com/belgium/east-flanders
//     benelux.com/belgium/brabant
//     ...
//     benelux.com/netherlands/overview
//     benelux.com/netherlands/drenthe
//     benelux.com/netherlands/flevoland
//     ...
//
// All content is accessible from a left nav in the usual app layout:
//
//      overview     |                             {content outline
//      belgium      |                              here - title, h1,
//     [netherlands] |       {main page             h2, h3 etc.}
//      luxembourg   |      content here}
//
// Core UX rules:
//   * At all times exactly one of the left nav choices is selected.
//   * The main page shows content associated with that selection.
// This makes it obvious where you are, where you can go, and how to get back.
//
// The first item, in this case "overview" is the initial highlight.
// If one hits the domain without a REST path, one is redirected to
// /overview and that item is highlighted in the menu, and its
// content is shown.
//
// Items in the left nav either name content and show it when clicked, or
// they name sub-tutorials and expand sub-tutorial choices when clicked.
// In the latter case, the main content and the left nav highlighting
// _do not change_.  A second click hides the exposed sub-tutorial names.
//
// Only the name of a Lesson (a leaf) with content can 1) be highlighted,
// 2) change the main page content when clicked, and 3) serve at a meaningful
// REST address.  Everything else is a sub-tutorial, and only expands or hides
// its own appearance.
//
// By design, this scheme maps to this filesystem layout:
//
//     benelux/
//       01_history.md
//       02_economy.md
//       README.md
//       03_belgium/
//         01_tintin.md
//         02_beer.md
//         03_antwerp/
//           README.md
//           01_diamonds.md
//           ...
//         04_east-flanders.md
//         05_brabant.md
//         ...
//       04_netherlands/
//         README.md
//         01_drenthe.md
//         02_flevoland.md
//       ...
//
// Where, say README (a github name convention) is converted to "overview"
// by a file loader.
//
// The proposed command line to read and serve content is
//
//      mdrip --mode web /foo/benelux
// or
//      mdrip --mode web /foo/benelux/README.md
//
// i.e. the argument names either a directory or a file.
//
// If the arg is a directory name, the tree below it is read in an attempt
// to build RESTfully addressable content and UX.  The names shown in the UX
// could be raw file names or could be processed a bit, e.g. underscores or
// hyphens become spaces, the ordering of the content in the UX could be
// controlled by omittable numerical prefixes on file names, etc.
// Errors in tree structure dealt with reasonably or cause immediate server
// failure.
//
// If only one file is read, then only that content is shown -
// no left nav needed.

type Tutorial interface {
	Name() string
	Path() model.FilePath
	// The order matters.
	Children() []Tutorial
	Accept(v Visitor)
}

type Visitor interface {
	VisitTopCourse(t *TopCourse)
	VisitCourse(c *Course)
	VisitLesson(l *Lesson)
	VisitCommandBlock(b *CommandBlock)
}

// A TopCourse is a Course with no name - it's the root of the tree (benelux).
type TopCourse struct {
	path     model.FilePath
	children []Tutorial
}

func NewTopCourse(p model.FilePath, c []Tutorial) *TopCourse { return &TopCourse{p, c} }
func (t *TopCourse) Accept(v Visitor)                        { v.VisitTopCourse(t) }
func (t *TopCourse) Name() string                            { return "" }
func (t *TopCourse) Path() model.FilePath                    { return t.path }
func (t *TopCourse) Children() []Tutorial                    { return t.children }

// A Course, or directory, has a name but no content, and an ordered list of
// Lessons and Courses. If the list is empty, the Course is dropped (hah!).
type Course struct {
	patrh    model.FilePath
	children []Tutorial
}

func NewCourse(p model.FilePath, c []Tutorial) *Course { return &Course{p, c} }
func (c *Course) Accept(v Visitor)                     { v.VisitCourse(c) }
func (c *Course) Name() string                         { return c.patrh.Base() }
func (c *Course) Path() model.FilePath                 { return c.patrh }
func (c *Course) Children() []Tutorial                 { return c.children }

// A Lesson, or file, must have a name, and should have blocks.
type Lesson struct {
	path      model.FilePath
	structure map[model.Label][]*CommandBlock
}

func convert(m map[model.Label][]*model.OldBlock) map[model.Label][]*CommandBlock {
	result := make(map[model.Label][]*CommandBlock)
	for k, v := range m {
		bar := []*CommandBlock{}
		for _, z := range v {
			bar = append(bar, NewCommandBlock(z.Labels(), z.RawProse(), z.Code()))
		}
		result[k] = bar
	}
	return result
}

func NewLesson(p model.FilePath, m map[model.Label][]*model.OldBlock) *Lesson {
	return &Lesson{p, convert(m)}
}

func (l *Lesson) Accept(v Visitor)                           { v.VisitLesson(l) }
func (l *Lesson) Name() string                               { return l.path.Base() }
func (l *Lesson) Path() model.FilePath                       { return l.path }
func (l *Lesson) Structure() map[model.Label][]*CommandBlock { return l.structure }
func (l *Lesson) Children() []Tutorial {
	result := []Tutorial{}
	for _, z := range l.structure[model.AnyLabel] {
		result = append(result, z)
	}
	return result
}

// CommandBlock groups opaqueCode with its labels.
type CommandBlock struct {
	labels []model.Label
	// prose is human language documentation for the opaqueCode
	prose []byte
	code  model.OpaqueCode
}

func NewCommandBlock(labels []model.Label, prose []byte, code model.OpaqueCode) *CommandBlock {
	if len(labels) < 1 {
		// Assure at least one label.
		labels = []model.Label{model.MistakeLabel}
	}
	return &CommandBlock{labels, prose, code}
}

func (x *CommandBlock) Accept(v Visitor)       { v.VisitCommandBlock(x) }
func (x *CommandBlock) Name() string           { return string(x.labels[0]) }
func (x *CommandBlock) Path() model.FilePath   { return model.FilePath("wutwutwut") }
func (x *CommandBlock) Labels() []model.Label  { return x.labels }
func (x *CommandBlock) Code() model.OpaqueCode { return x.code }
func (x *CommandBlock) Children() []Tutorial   { return []Tutorial{} }
func (x *CommandBlock) RawProse() []byte       { return x.prose }
func (x *CommandBlock) Prose() template.HTML {
	return template.HTML(string(blackfriday.MarkdownCommon(x.prose)))
}

func (x *CommandBlock) Print(
	w io.Writer, prefix string, n int, label model.Label, fileName model.FilePath) {
	fmt.Fprintf(w, "echo \"%s @%s (block #%d in %s) of %s\"\n\n",
		prefix, x.Name(), n, label, fileName)
	fmt.Fprint(w, x.Code())
}
