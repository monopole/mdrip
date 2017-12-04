# mdrip

[fenced code blocks]: https://help.github.com/articles/creating-and-highlighting-code-blocks/#fenced-code-blocks
[travis-mdrip]: https://travis-ci.org/monopole/mdrip
[tmux]: https://github.com/tmux/tmux/wiki

`mdrip` turns a directory hierarchy of markdown into a navigable, executable, testable tutorial.

* Directory structure becomes hierarchical navigation.
* Navigate through all content with arrow keys.
* Track progress with check marks.
* Click on code block headers to immediately execute them in [tmux].

Or just use `mdrip` to place markdown-based tutorials under CI/CD testing,
to avoid disappointing users.



[![Build Status](https://travis-ci.org/monopole/mdrip.svg?branch=master)](https://travis-ci.org/monopole/mdrip)


## Installation

Assuming [Go](https://golang.org/dl) installed just:

```
go get github.com/monopole/mdrip
```

or put it in your tmp dir
```
GOBIN=$TMPDIR go install github.com/monopole/mdrip
```

## Execution

> `mdrip {filePath}`

This searches the given path for files named
`*.md` (ignoring everything else), and parses
the markdown into memory.

The `filePath` argument can be

* a single local file,
* a local directory,
* a github URL in the style `gh:{handle}/{repoName}`,
* or a particular file or a directory in the repo, e.g. `gh:{handle}/{repoName}/foo/bar`.

What happens next depends on the `--mode` flag.

## Demo Mode: make a tutorial web app

> `mdrip --mode demo {filePath}`

This serves rendered markdown at
`http://localhost:8000`.  Change the endpoint using
`--port` and `--hostname`.

Clicking on a code block in your browser will
copy its contents to your clipboard, and if
you happen to have a local instance of [tmux]
running, the `mdrip` server sends the code
block directly to the currently active tmux
pane for immediate execution.

This one-click operation is handy for demos.

##### Example:

Render the content you are now reading locally:
```
$TMPDIR/mdrip --mode demo gh:monopole/mdrip/README.md
```

Visit [localhost:8000](http://localhost:8000).

If you have it, start tmux.  Click on command
blocks in your browser to send them directly
to your active tmux window.


## Print Mode: extract code to stdout

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

##### Example:

Print all the fenced code blocks from this `README`:

```
$TMPDIR/mdrip gh:monopole/mdrip/README.md
```

## Test Mode: place markdown code under test

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

#### Special labels

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


#### Example:

[Go tutorial]: https://github.com/monopole/mdrip/blob/master/data/example_tutorial.md
[raw-example]: https://raw.githubusercontent.com/monopole/mdrip/master/data/example_tutorial.md

This [Go tutorial] has code blocks that write, compile
and run a Go program.

Use this to extract blocks to `stdout`:

```
alias mdrip=$TMPDIR/mdrip
mdrip --label lesson1 \
    gh:monopole/mdrip/data/example_tutorial.md
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


So, adding a line like

```
mdrip --mode test --label {someLabel} {filePath}
```
to your CI/CD test framework covers
the execution path determined by that label.


## Tips for writing markdown tutorials

[fenced code blocks]: https://help.github.com/articles/creating-and-highlighting-code-blocks/#fenced-code-blocks
[block quote]: https://github.github.com/gfm/#block-quotes

 * Place commands that the reader should copy/paste/execute in
   [fenced code blocks].

 * Eschew preceding commands with fake prompts (e.g. `$`).
   They are redundant, and complicate copy/paste.

 * Code-style text not intended for immediate execution, e.g. alternative
   commands or example output, should be in a fenced code block indented via a
   [block quote].  This makes them invisible to `mdrip`.
