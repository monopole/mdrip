// For testing there are two assumptions.
// * At the time of writing, we want to use the same files for two Markdown parsers
//   https://github.com/yuin/goldmark
//   https://github.com/gomarkdown/markdown
// * The *_test.go files, when running, want to read test data from files in or
//   below the directory in which the *_test.go files live.
// To accommodate these assumptions, the following generate directive
// creates a symlink to the test data.
// (disabledgo):generate ln -is ../usegold/testdata .

package useblue_test
