package program

import "errors"

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
// At all times _one_ of the left nav choices is selected and highlighted,
// and the main page shows content associated with that selection.  That's
// a core interaction - the content shown is known to be associated with a
// highlighted element in the left nav.  It's obvious how to get back to it
// if something else is clicked.
//
// The overview is the initial highlight.  If one hits the domain without a
// REST path, one is redirected to /overview and that item is highlighted in
// the menu, and its content is shown.
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
// If only one file is read, then only that is shown - no left nav needed.

// Key data structure for the tree, used to build three things:
//
// A Course, or directory, has a name, no content, and at least one child

type Tutorial interface {
	Name() string
	Content() string
	// The order matters.
	Children() []Tutorial
}

// A Lesson, or file, must have a name, must have content and zero children.
type Lesson struct {
	name string
	content string
}
func (l Lesson) Name() string { return l.name }
func (l Lesson) Content() string { return l.content }
func (l Lesson) Children() []Tutorial { return []Tutorial{} }

// A Course, or directory, has a name, no content, and an ordered list of
// Lessons and Courses. If the list is empty, the Course is dropped.
type Course struct {
	name string
	children []Tutorial
}
func (c Course) Name() string { return c.name }
func (c Course) Content() string { return "" }
func (c Course) Children() []Tutorial { return c.children }

// A TopCourse is a Course with no name - it's the root of the tree (benelux).
type TopCourse struct {
	children []Tutorial
}
func (t TopCourse) Name() string { return "" }
func (t TopCourse) Content() string { return "" }
func (t TopCourse) Children() []Tutorial { return t.children }

func isDirectory(name string) bool { return true }
func isTextFile(name string) bool { return true }

// Load loads a tutorial tree from disk.
func Load(name string) (Tutorial, error) {
	if isDirectory(name) {
		return TopCourse{[]Tutorial{}}, nil
	}
	if isTextFile(name) {
		return Lesson{name, "some content"}, nil
	}
	return nil, errors.New("arg is neither file or directory")
}
