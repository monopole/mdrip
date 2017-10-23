# mdrip

[fenced code blocks]: https://help.github.com/articles/github-flavored-markdown/#fenced-code-blocks
[_here_ documents]: http://tldp.org/LDP/abs/html/here-docs.html
[literate programming]: http://en.wikipedia.org/wiki/Literate_programming
[tmux]: https://github.com/tmux/tmux/wiki
[travis-mdrip]: https://travis-ci.org/monopole/mdrip
[Go tutorial]: https://github.com/monopole/mdrip/blob/master/data/example_tutorial.md
[raw-example]: https://raw.githubusercontent.com/monopole/mdrip/master/data/example_tutorial.md

[![Build Status](https://travis-ci.org/monopole/mdrip.svg?branch=master)](https://travis-ci.org/monopole/mdrip)

`mdrip` rips [fenced code blocks] from markdown files,
making them available for execution in
[_demo_](#demo-mode),
[_print_](#print-mode) and
[_test_](#test-mode)
modes.

## Installation

Assuming Go installed:

```
export MDRIP_HOME=$(mktemp -d)
GOPATH=$MDRIP_HOME go get github.com/monopole/mdrip
alias mdrip=$MDRIP_HOME/bin/mdrip
```

or just
```
go get github.com/monopole/mdrip
```

## Execution modes

`mdrip` has various flags, but accepts only one bare argument:

> `mdrip {filePath}`

The program searches the given path for files named
`*.md`, and parses the markdown into memory.  What
happens next depends on the `--mode` flag.

The `filePath` argument can be a file name, a directory
path, or a github URL in the style
`gh:{handle}/{repoName}`.  The last case is a convenience
that clones the repo into a disposable tmp dir and scans
its contents from there in one step.

### demo mode

This mode faciliates markdown-based demos.

```
mdrip --mode demo {filePath}
```

serves rendered markdown at `http://localhost:8000`
(see also flags `--port` and `--hostname`).

Clicking on a code block in your browser will
copy its contents to your clipboard, and if you happen
to have a local instance of [tmux] running, the `mdrip`
server also sends the code block directly to the
currently active tmux pane as if it had been manually
pasted.

This one-click operation is surprisingly handy for
demos wherein one has a tmux window next to a browser
window.


### print mode

In this default mode, extracted code blocks are printed to `stdout`.

```
eval "$(mdrip file.md)"
```
runs extracted blocks in the current terminal, while
```
mdrip file.md | source /dev/stdin
```
runs extracted blocks in a piped shell that exits with extracted code status.

The difference between these two mode of operation is the
same as the difference between
`eval "$( exit )"` and `echo exit | source /dev/stdin`.
The former affects your terminal, the latter does not.

To stop on error, pipe `mdrip` output to `bash -e`.

### test mode

To assure that, say, a tutorial about some procedure
continues to work, some test suite can assert that the
following command exits with status 0:

```
mdrip --mode test /path/to/tutorial.md
```

This runs extracted blocks in an `mdrip` subshell,
leaving the executing shell unchanged.

In this mode, `mdrip` captures the stdout and stderr of
the subprocess, reporting only blocks that fail,
facilitating error diagnosis.  Normally, mdrip exits
with non-zero status only when used incorrectly,
e.g. file not found, bad flags, etc.  In in test mode,
mdrip will exit with the status of any failing code
block.

This mode is an instance of [literate programming] in
that code (shell commands) are embedded in explanatory
content (markdown).

#### Labels

Fenced code blocks can be preceeded in the markdown by
a one-line HTML comment with embedded labels in this form:

<blockquote>
<pre>
&lt;&#33;-- @initializeCluster @tutorial03 @test --&gt;
&#96;&#96;&#96;
echo hello
&#96;&#96;&#96;
</pre>
</blockquote>

Then the `--label` flag can be used to extract only
code blocks with the given label, e.g.

> `mdrip --label test {filePath}`

discards all code blocks other than those with a
preceding `@test` label.

Use this to ignore blocks that aren't suitable for a
test sequence, e.g. a sequence that
password-authenticates a particular user (a test
framework will do this by other means).

##### Special labels

 * The first label on a block is slightly special, in
   that it's reported as the block name for logging in
   test mode.  So its useful to treat it as a method
   name for the block, e.g. `@initializeCluster` or
   `@createWebServer`.

 * The label `@sleep` causes mdrip to insert a `sleep
   2` command _after_ the block.  Appropriate if one is
   starting a server in the background in that block,
   and want to impose a short wait (which you'd get
   implicitly if a human were executing the blocks more
   slowly as part of a demo).


## Example

This short [Go tutorial] has bash code blocks that
write, compile and run a Go program.

Send code from that file to `stdout`:

```
mdrip lesson1 \
    $MDRIP_HOME/src/github.com/monopole/mdrip/data/example_tutorial.md
```

Alternatively, run its code in a subshell:
```
mdrip --mode test lesson1 \
    $MDRIP_HOME/src/github.com/monopole/mdrip/data/example_tutorial.md
```

The above command has no output and exits with status zero if all the
scripts labelled `@lesson1` in the given markdown succeed.  On any
failure, however, the command dumps a report and exits with non-zero
status.
