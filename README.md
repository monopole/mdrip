# mdrip

[fenced code blocks]: https://help.github.com/articles/creating-and-highlighting-code-blocks/#fenced-code-blocks
[travis-mdrip]: https://travis-ci.org/monopole/mdrip
[tmux]: https://github.com/tmux/tmux/wiki

Rips code blocks from markdown and makes it useful.

[![Build Status](https://travis-ci.org/monopole/mdrip.svg?branch=master)](https://travis-ci.org/monopole/mdrip)
[![Go Report Card](https://goreportcard.com/badge/github.com/monopole/mdrip)](https://goreportcard.com/report/github.com/monopole/mdrip)

### Testing

Extract and run all code block in all the markdown 
in and below your current directory:
```
mdrip test
```
The above command exits successfully of all the code blocks do, else it exits 
with an error code.

Add this to your CI to assure that the code in your markdown works.

It's unlikely that you want to run every block;
use a [label](#labels) to be selective.

### Give presentations

To convert your markdown into an interactive tutorial 
that works with [tmux], run
```
mdrip demo
```
You can then run blocks in a tmux session with keys strokes only (and no copy/paste).


## Installation

Assuming [Go](https://golang.org/dl) installed just:

```
GOBIN=$TMPDIR go install github.com/monopole/mdrip/v2
```

## The Details

### Printing and running

> `mdrip print {filePath}`

This searches the given path for files named
`*.md`, parses the markdown into memory, then
emits code blocks as one script.

The `filePath` argument can be

* a single local file,
* a local directory,
* a github URL in the style `gh:{user}/{repoName}`,
* or a particular file or a directory in the repo, e.g. `gh:{user}/{repoName}/foo/bar`.

To extract and noisily run blocks in the current terminal:

> `eval "$(mdrip print goTutorial.md)"`

It's better to pipe them to a subprocess:

> `mdrip goTutorial.md | source /dev/stdin`

The difference between these two compositions is the
same as the difference between

> `eval "$(exit)"`

and

> `echo exit | source /dev/stdin`

The former affects your current shell, the latter doesn't.

To stop on error, pipe `mdrip` output to `bash -e`.

Te get better reporting on which blocks fail, use the `test`
command:

> `mdrip test goTutorial.md`

The stdout and stderr of the subprocess are captured,
an only the output associated with a failing block
is reported.

### Labels

One can _label_ a code block by preceding it with
a one-line HTML comment, e.g:

<blockquote>
<pre>
&lt;&#33;-- @initializeCluster @test  @tutorial03 --&gt;
&#96;&#96;&#96;
echo hello
&#96;&#96;&#96;
</pre>
</blockquote>

One can then use the `--label` flag to select only
code blocks with that label, e.g.

> `mdrip test --label test {filePath}`

The first label on a block is slightly special, in
that it's reported as the block's _name_ for various
purposes.  If no labels are present, one is generated
for these purposes.

[literate programming]: http://en.wikipedia.org/wiki/Literate_programming
[_here_ documents]: http://tldp.org/LDP/abs/html/here-docs.html

This mode is an instance of [literate programming] in
that code (shell commands) are embedded in explanatory
content (markdown).  One can use [_here_ documents] to
incorporate any programming language into the tests
(as in [goTutorial.md](./goTutorial.md) below).

### Debugging and demonstrations

The command

> `mdrip demo goTutorial.md`

serves rendered markdown at `http://localhost:8000`.

Hit '?' in the browser to see key controls.

If you have a local instance of [tmux]
running, the `mdrip` server sends the code
block directly to active tmux
pane for immediate execution.

#### Example:

[Go tutorial]: https://github.com/monopole/mdrip/blob/master/data/example_tutorial.md
[raw-example]: https://raw.githubusercontent.com/monopole/mdrip/master/data/example_tutorial.md

This [Go tutorial] has code blocks that write, compile
and run a Go program.

Use this to extract blocks to `stdout`:

```
alias mdrip=$TMPDIR/mdrip
mdrip --label lesson1 gh:monopole/mdrip/goTutorial.md
```

Test the code from the markdown in a subshell:
```
clear
mdrip --mode test --label lesson1 gh:monopole/mdrip/goTutorial.md
echo $?
```

The above command should show an error, and exit with non-zero status,
because that example tutorial has several baked-in errors.

To see success, download the example and confirm
that it fails locally:
```
tmp=$(mktemp -d)
git clone https://github.com/monopole/mdrip.git $tmp
file=$tmp/goTutorial.md
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
mdrip test --label {someLabel} {filePath}
```
to your CI/CD test framework covers
the execution path determined by that label.


## Tips for writing markdown tutorials

[fenced code blocks]: https://help.github.com/articles/creating-and-highlighting-code-blocks/#fenced-code-blocks
[block quote]: https://github.github.com/gfm/#block-quotes

Place commands that the reader should copy/paste/execute in
[fenced code blocks].

Code-style text not intended for immediate execution, e.g. alternative
commands or example output, should be in a fenced code block indented via a
[block quote].  This makes them invisible to `mdrip`.

Don't put prompts in your code blocks.

The following is easy to copy/paste:
> ```
> echo hello
> du -sk
> ```
But this isn't:
> ```
> $ echo hello
> $ du -sk
> ```

