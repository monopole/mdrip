package program

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/golang/glog"
)

// Tutorial UX Overview.
//
// Suppose it's a tutorial on Benelux.
//
// The first lesson is an overview of Benelux, with sibling (not child) lessons
// covering Belgium, Netherlands, and Luxembourg.  These in turn could contain
// lessons on provinces, which could contain lessons on cities, etc.
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
// Crucially, all content is accessible from a left nav in a page like this:
//
//      overview     |                             {content outline
//      belgium      |                              here - title, h1,
//     [netherlands] |       {main page             h2, h3 etc.}
//      luxembourg   |      content here}
//
// The core interaction here is that
//   * At all times exactly one of the left nav choices is selected.
//   * The main page shows content associated with that selection.
// It's always obvious where you are, where you can go, and how to get back.
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
	Content() string
	// The order matters.
	Children() []Tutorial
	Print(indent int)
}

// A Lesson, or file, must have a name, must have content and zero children.
type Lesson struct {
	name    string
	content string
}

func filterNewLines(s string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case 0x000A, 0x000B, 0x000C, 0x000D, 0x0085, 0x2028, 0x2029:
			return ' '
		default:
			return r
		}
	}, s)
}

const maxSummaryLen = 50

func (l *Lesson) Name() string         { return l.name }
func (l *Lesson) Content() string      { return l.content }
func (l *Lesson) Children() []Tutorial { return []Tutorial{} }
func (l *Lesson) Print(indent int) {
	s := len(l.content)
	if s > maxSummaryLen {
		s = maxSummaryLen
	}
	z := strings.TrimSpace(l.content[:s])
	z = filterNewLines(z)
	fmt.Printf(spaces(indent)+"%s --- %s...\n", l.name, z)
}

// A Course, or directory, has a name, no content, and an ordered list of
// Lessons and Courses. If the list is empty, the Course is dropped.
type Course struct {
	name     string
	children []Tutorial
}

func spaces(indent int) string {
	if indent < 1 {
		return ""
	}
	return fmt.Sprintf("%"+strconv.Itoa(indent)+"s", " ")
}

func (c *Course) Name() string         { return c.name }
func (c *Course) Content() string      { return "" }
func (c *Course) Children() []Tutorial { return c.children }
func (c *Course) Print(indent int) {
	fmt.Printf(spaces(indent)+"%s\n", c.name)
	for _, x := range c.children {
		x.Print(indent + 3)
	}
}

// A TopCourse is a Course with no name - it's the root of the tree (benelux).
type TopCourse struct {
	children []Tutorial
}

func (t *TopCourse) Name() string         { return "" }
func (t *TopCourse) Content() string      { return "" }
func (t *TopCourse) Children() []Tutorial { return t.children }
func (t *TopCourse) Print(indent int) {
	for _, x := range t.children {
		x.Print(indent)
	}
}

const badLeadingChar = "~.#"

func isDesirableFile(n string) bool {
	s, err := os.Stat(n)
	if err != nil {
		glog.Info("Stat error on "+s.Name(), err)
		return false
	}
	if s.IsDir() {
		glog.Info("Ignoring NON-file " + s.Name())
		return false
	}
	if !s.Mode().IsRegular() {
		glog.Info("Ignoring irregular file " + s.Name())
		return false
	}
	if filepath.Ext(s.Name()) != ".md" {
		glog.Info("Ignoring non markdown file " + s.Name())
		return false
	}
	base := filepath.Base(s.Name())
	if strings.Index(badLeadingChar, string(base[0])) > -1 {
		glog.Info("Ignoring because bad leading char: " + s.Name())
		return false
	}
	return true
}

func isDesirableDir(n string) bool {
	s, err := os.Stat(n)
	if err != nil {
		glog.Info("Stat error on "+s.Name(), err)
		return false
	}
	if !s.IsDir() {
		glog.Info("Ignoring NON-dir " + s.Name())
		return false
	}
	if s.Name() == "." || s.Name() == "./" || s.Name() == ".." {
		// Allow special names.
		return true
	}
	if strings.HasPrefix(filepath.Base(s.Name()), ".") {
		glog.Info("Ignoring dot dir " + s.Name())
		// Ignore .git, etc.
		return false
	}
	return true
}

func scanDir(d string) (*Course, error) {
	files, err := ioutil.ReadDir(d)
	if err != nil {
		return nil, err
	}
	var items = []Tutorial{}
	for _, f := range files {
		p := filepath.Join(d, f.Name())
		if isDesirableFile(p) {
			l, err := scanFile(p)
			if err != nil {
				return nil, err
			}
			items = append(items, l)
		} else if isDesirableDir(p) {
			c, err := scanDir(p)
			if err != nil {
				return nil, err
			}
			if c != nil {
				items = append(items, c)
			}
		}
	}
	if len(items) > 0 {
		return &Course{filepath.Base(d), items}, nil
	}
	return nil, nil
}

func scanFile(n string) (*Lesson, error) {
	contents, err := ioutil.ReadFile(n)
	if err != nil {
		return nil, err
	}
	return &Lesson{filepath.Base(n), string(contents)}, nil
}

func Load(root string) (Tutorial, error) {
	if isDesirableFile(root) {
		return scanFile(root)
	}
	if isDesirableDir(root) {
		c, err := scanDir(root)
		if err != nil {
			return nil, err
		}
		if c != nil {
			return &TopCourse{c.children}, nil
		}
	}
	return nil, errors.New("Cannot process " + root)
}
