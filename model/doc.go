package model

// Package model has types used to build a tutorial from discovered markdown.
//
// The file hierarchy holding the markdown defines tutorial structure,
// organizing markdown files (lessons) into nestable groups (courses).
//
// Example: tutorial on Benelux.
//
// The first lesson could be an overview of Benelux, with sibling (not child)
// courses covering Belgium, Netherlands, and Luxembourg - as one would begin
// a textbook with an introduction.
//
// Said courses may hold lessons on provinces, or sub-courses regional
// histories, cities etc.  A user could drop in anywhere, but content should
// be arranged such that a depth-first traversal of the hierarchy is a
// meaningful path through all content - i.e. one ought to be able to read
// it as a book.
//
// Associated content REST addresses reflect file system hierarchy, e.g.
//
//     benelux.com/overview                  // Describes Benelux in general.
//     benelux.com/history                   // Benelux history, economy, etc.
//     benelux.com/economy
//     benelux.com/belgium/overview          // Describes Belgium in general.
//     benelux.com/belgium/tintin            // Dive into details.
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
// All content is accessible from a left nav:
//
//      overview     |                           {main page outline
//      belgium      |                            here - title, h1,
//     [netherlands] |       {main page           h2, h3 etc.}
//      luxembourg   |        content here}
//
// At all times exactly one of the left nav choices is selected, and the
// main page shows content associated with that selection.
//
// The first item, in this case the "overview", is the initial highlight.
// If one hits the domain without a REST path, one is redirected to
// /overview and that item is highlighted in the menu, and its
// content is shown.
//
// Items in the left nav either name content and show it when clicked, or
// they name sub-courses and expand choices when clicked.
// In the latter case, the main content and the left nav highlighting
// don't change.  Subsequent clicks toggle sub-course names.
//
// Only the name of a lesson (a leaf) with content can 1) be highlighted,
// 2) change the main page content when clicked, and 3) serve at a meaningful
// REST address.  Everything else is a course, and only expands or hides
// its own appearance.
//
// This scheme maps to this filesystem layout:
//
//     benelux/
//       history.md
//       economy.md
//       README.md
//       belgium/
//         tintin.md
//         beer.md
//         antwerp/
//           README.md
//           diamonds.md
//           ...
//         east-flanders.md
//         brabant.md
//         ...
//       netherlands/
//         README.md
//         drenthe.md
//         flevoland.md
//       ...
//
// Where, say README is converted to "overview" by a file loader, and likewise
// the ordering of files and directories in the tutorials is presented in, say,
// a file called README_ORDER.txt so that 'history' precedes 'economy', etc.
//
// Useful data structures would facilitate mapping a string path, e.g.
//   belgium/antwerp/diamonds
// to some list of div ids to know what items to show/hide when first
// loading a page.
//
// Another handy structure would facilitate a prev/next navigation through
// the lessons.  The lessons are all leaves of the tree.  Moving from lesson
// n to lesson n+1 means changing the main content and changing what is shown
// or highlighted in the left nav.