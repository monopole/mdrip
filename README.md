# mdrip

[fenced code blocks]: https://help.github.com/articles/creating-and-highlighting-code-blocks/#fenced-code-blocks
[travis-mdrip]: https://travis-ci.org/monopole/mdrip
[tmux]: https://github.com/tmux/tmux/wiki

Rips code blocks from markdown and makes them useful.

[![Build Status](https://travis-ci.org/monopole/mdrip.svg?branch=master)](https://travis-ci.org/monopole/mdrip)
[![Go Report Card](https://goreportcard.com/badge/github.com/monopole/mdrip)](https://goreportcard.com/report/github.com/monopole/mdrip)

### Testing

Extract and run all code block in all the markdown 
in and below your current directory:
> ```
> mdrip test
> ```

This exits successfully if all the commands in the code blocks do,
else it exits with the error code of the failed command.

Add this to your CI to assure that the code in your markdown works,
assuming it's written to do so.

It's unlikely that you'll want to run every block;
use a [label](#labels) to be selective.

### Give presentations

To convert your markdown into an interactive tutorial 
that works with [tmux], run

> ```
> mdrip demo
> ```

You can then send blocks to a tmux session with keys strokes (no copy/paste).


## Installation

Assuming [Go](https://golang.org/dl) installed just:

```
go install github.com/monopole/mdrip/v2
```

## The Details

To run the following examples, generate some markdown locally using
```
mdrip gentestdata mdTestData
```
The final argument is the name of a directory to create and fill with markdown.

This test data has code blocks with read-only commands, e.g. `cat /etc/hosts`.

Also grab the `brokenGoTutorial.md` file:
```
curl -O https://raw.githubusercontent.com/monopole/mdrip/master/brokenGoTutorial.md
```
The code blocks in this file will create some files in your `TMPDIR`.

Obviously, inspect as desired.

### Printing and running

The sub-command `print` searches the given path for `*.md`,
parses the markdown into memory, then emits code blocks as one script.

```
clear
mdrip print mdTestData | head -n 40
```

The argument to `print` can be

* a single local file,
* a local directory,
* a github URL in the style `gh:{user}/{repoName}`,
* or a particular file or a directory in the repo, e.g. `gh:{user}/{repoName}/foo/bar`.

So one can pipe the blocks into a subprocess with:
```
clear
mdrip print mdTestData/dir5 | source /dev/stdin
```
Or send the blocks to a subprocess that stops on the first error:
```
clear
mdrip print mdTestData/dir5 | bash -e
```
Or run them in your current shell (handy for setting env variables that you wish to use):
```
clear
eval "$(mdrip print mdTestData/dir5)"
```

To get better reporting on which blocks fail, use the `test`
command:

```
clear
mdrip test brokenGoTutorial.md
```

The stdout and stderr of the subprocess are captured,
an only the output associated with a failing block
is reported.  This is better for use in CI/CD situations,
since the failing code block will be easier to spot.

### Labels

One can _label_ a code block by preceding it with
a one-line HTML comment, e.g:

<blockquote>
<pre>
&lt;&#33;-- @initializeCluster @test @tutorial03 --&gt;
&#96;&#96;&#96;
echo hello
&#96;&#96;&#96;
</pre>
</blockquote>

Labels are just words that start with an `@` in the comment.

One can then use the `--label` flag to select only
code blocks with that label, e.g.

```
clear
mdrip print --label mississippi mdTestData/dir2 | head -n 40
```

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
(as in [brokenGoTutorial.md](./brokenGoTutorial.md) below).

### Debugging and demonstrations

The command
```
mdrip demo mdTestData
```
serves rendered markdown at `http://localhost:8000`.

Hit `?` in the browser to see key controls.

If you have a local instance of [tmux]
running, the `mdrip` server sends the code
block directly to active tmux
pane for immediate execution.

#### Example:

[Go tutorial]: https://github.com/monopole/mdrip/blob/master/brokenGoTutorial.md
[raw-example]: https://raw.githubusercontent.com/monopole/mdrip/master/brokenGoTutorial.md

This [Go tutorial] has code blocks that write, compile
and run a Go program.

Use this to extract blocks to `stdout`:

```
clear
mdrip print --label lesson1 brokenGoTutorial.md
```

Test the code from the markdown in a subshell:
```
clear
mdrip test --label lesson1 brokenGoTutorial.md
echo $?
```

The above command should show an error, and exit with non-zero status,
because that example tutorial has several baked-in errors.

Fix the problems:
```
cp brokenGoTutorial.md goTutorial.md
sed -i 's|comment this|// comment this|' goTutorial.md
sed -i 's|intended to fail|intended to succeed|' goTutorial.md
sed -i 's|badCommandToTriggerTestFailure|echo Hello|' goTutorial.md
echo "  "
diff brokenGoTutorial.md goTutorial.md 
```

Run the test again:
```
clear
mdrip test --label lesson1 goTutorial.md
echo $?
```

The return code should be zero.

So, adding a line like

> ```
> mdrip test --label {someLabel} {filePath}
> ```

to your CI/CD test framework covers
the execution path determined by that label.


## Tips for writing markdown tutorials

[fenced code blocks]: https://help.github.com/articles/creating-and-highlighting-code-blocks/#fenced-code-blocks
[block quote]: https://github.github.com/gfm/#block-quotes

Place commands that the reader would want to execute directly
(with no edits) in
[fenced code blocks].

Code-style text _not intended_ for copy/paste execution, e.g. alternative
commands with fake arguments, or example code or output,
should be in a fenced code block indented via a
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

