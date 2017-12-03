# mdrip

[fenced code blocks]: https://help.github.com/articles/creating-and-highlighting-code-blocks/#fenced-code-blocks
[block quote]: https://github.github.com/gfm/#block-quotes
[travis-mdrip]: https://travis-ci.org/monopole/mdrip

`mdrip` turns a directory hierarchy of markdown into an ordered, book-like
tutorial.

It rips [fenced code blocks] from markdown files,
making them available for execution in tests and
demos.

In test mode, `mdrip` runs particular code blocks and reports
failures in execution.

In demo mode, `mdrip` renders the markdown as
a tutorial allowing the user to:

* Navigate through a directory hierarchy via arrow keys.
* Click on code block headers to immediately execute them in tmux.
* Track progress with check marks.


[![Build Status](https://travis-ci.org/monopole/mdrip.svg?branch=master)](https://travis-ci.org/monopole/mdrip)


## Installation

Assuming [Go](https://golang.org/dl) installed just:

```
go get github.com/monopole/mdrip
```

or put it in your tmp dir
```
GOBIN=$TMPDIR go install github.com/monopole/mdrip
alias mdrip=$TMPDIR/mdrip
```

## Execution

> `mdrip {filePath}`

searches the given path for files named `*.md`, and
parses the markdown into memory.

The `filePath` argument can be a single file, a
directory, or a github URL in the style
`gh:{handle}/{repoName}`.  This last is a convenience
that clones the repo into a disposable `/tmp` directory
and scans the markdown from there.

What happens next depends on the `--mode` flag.

## Modes

### demo: facilitate markdown-based demos

> `mdrip --mode demo {filepath or github URL}`

This serves rendered markdown at
`http://localhost:8000`.  Change the endpoint using
`--port` and `--hostname`.

[tmux]: https://github.com/tmux/tmux/wiki

Clicking on a code block in your browser will
copy its contents to your clipboard, and if you happen
to have a local instance of [tmux] running, the `mdrip`
server also sends the code block directly to the
currently active tmux pane as if it had been manually
pasted.

This one-click operation is surprisingly handy for
demos wherein one has a tmux window next to a browser
window.

##### example:

Render the content you are now reading locally:
```
GOBIN=$TMPDIR go install github.com/monopole/mdrip
$TMPDIR/mdrip --mode demo gh:monopole/mdrip/README.md
```

Visit [localhost:8000](http://localhost:8000).
If you have it, start tmux.  Click on command blocks
in your browser to send them
directly to your active tmux window.


### print: extract code to stdout

In this default mode, the command

> `eval "$(mdrip file.md)"`

runs extracted blocks in the current
terminal, while

> `mdrip file.md | source /dev/stdin`

runs extracted blocks in a piped shell that exits with
extracted code status.  The difference between these
two compositions is the same as the difference between

> `eval "$(exit)"`

and

> `echo exit | source /dev/stdin`

The former affects your current shell, the latter does
not.  To stop on error, pipe `mdrip` output to `bash
-e`.

### test: place markdown code under test

> `mdrip --mode test /path/to/tutorial.md`

runs extracted blocks in an `mdrip` subshell,
leaving the executing shell unchanged.

In this mode, `mdrip` captures the stdout and stderr of
the subprocess, reporting only blocks that fail,
facilitating error diagnosis.  Normally, mdrip exits
with non-zero status only when used incorrectly,
e.g. file not found, bad flags, etc.  In test mode,
mdrip will exit with the status of any failing code
block.

[literate programming]: http://en.wikipedia.org/wiki/Literate_programming
[_here_ documents]: http://tldp.org/LDP/abs/html/here-docs.html

This mode is an instance of [literate programming] in
that code (shell commands) are embedded in explanatory
content (markdown).  One can use [_here_ documents] to
incorporate any programming language into the tests
(see the [example](#example) below).

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

This can be used, for example, to gather blocks that
should be placed under test, and ignore those that
shouldn't.  An example of the latter would be commands
that prompt an interactive user to login (test
frameworks typically have their own notion of an
authenticated user).

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

[Go tutorial]: https://github.com/monopole/mdrip/blob/master/data/example_tutorial.md
[raw-example]: https://raw.githubusercontent.com/monopole/mdrip/master/data/example_tutorial.md

This [Go tutorial] has code blocks that write, compile
and run a Go program.

Use this to extract blocks to `stdout`:

```
mdrip --label lesson1 gh:monopole/mdrip/data/example_tutorial.md
```

Test the code from the markdown in a subshell:
```
clear
mdrip --mode test --label lesson1 \
    gh:monopole/mdrip/data/example_tutorial.md
echo $?
```

The above command should show an error, and exit with non-zero status,
because that example tutorial has several baked-in errors.

To see success, download the example and confirm
that it fails locally:
```
tmp=$(mktemp -d)
git clone https://github.com/monopole/mdrip.git $tmp
file=$tmp/data/example_tutorial.md
mdrip --mode test --label lesson1 $file
```

Fix the problems:
```
sed -i 's|comment this|// comment this|' $file
sed -i 's|intended to fail|intended to succeed|' $file
sed -i 's|badCommandToTriggerTestFailure|echo Hello|' $file
```

Run the test again:
```
mdrip --mode test --label lesson1 $file
echo $?
```

The return code should be zero.

## Tips for writing markdown tutorials

 * Place commands that the reader should copy/paste/execute in
   [fenced code blocks].

 * Eschew preceding commands with fake prompts (e.g. `$ ls`).
   They just complicate copy/paste.

 * Code-style text not intended for execution, e.g. example output
   or dangerous commands, should be in a fenced code block indented via a
   [block quote], e.g.
   > ```
   > rm -rf /
   > ```
