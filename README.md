
This tool makes markdown-based coding tutorials executable and
testable.  It's a hacky, markdown-based instance of
language-independent [literate
programming](http://en.wikipedia.org/wiki/Literate_programming) (for
perspective, see the latex-based
[noweb](http://en.wikipedia.org/wiki/Noweb)).

The tool scans markdown for [fenced code
blocks](https://help.github.com/articles/github-flavored-markdown/#fenced-code-blocks)
immediately preceded by an HTML comment with embedded _@labels_ and
extracts the labels and blocks.  The code blocks can then be piped to
an arbitrary interpreter.  The labels are used for block selection and
logging.

If the code blocks are in bash syntax, and the tool is itself running
in a bash shell, then piping `mdrip` output to `source /dev/stdin` is
equivalent to a human copy/pasting all the code blocks to their shell
prompt.

Alternatively, the tool can run extracted code in a bash subshell with
customizable parsing of the subshell's stdout and stderr, allowing
reporting like _block 'nesbit' from script 'foo' in file 'bar' failed with
error 'baz'_.  This behavior facilitates adding tutorial coverage to
regression test frameworks.

## Details

The tool has a simple extensibility because shell scripts can
make, build and run programs in any programming language, via [_here_
documents](http://tldp.org/LDP/abs/html/here-docs.html) and what not.

For example, this
[tutorial](https://github.com/monopole/mdrip/blob/master/example_tutorial.md)
(raw markdown
[here](https://raw.githubusercontent.com/monopole/mdrip/master/example_tutorial.md))
has bash code blocks that write, compile and run a Go program.

The tool accepts a _label_ argument and any number of _file name_
arguments then extracts all blocks with that label from those files,
retaining the block order.

A _script_ is a sequence of code blocks with a common label.  If a
block has multiple labels, it can be incorporated into multiple
scripts.  If a block has no label, it's ignored.  The number of
scripts that can be extracted from a set of markdown files equals the
number of unique labels.

The first label on a block is slightly special, in that it's
reported as the block name for logging.  But like any label
it can be used for selection too.

Beware that extracted scripts can do anything that the user can do.
There's no notion of encapsulation or automatic cleanup.  Blocks to do
clean can be added to the markdown.

## Build

Assuming Go installed:

```
export MDRIP=~/mdrip
GOPATH=$MDRIP/go go get github.com/monopole/mdrip
GOPATH=$MDRIP/go go test github.com/monopole/mdrip
$MDRIP/go/bin/mdrip   # Shows usage.
```

Send code from the [example tutorial]
(https://github.com/monopole/mdrip/blob/master/example_tutorial.md) to
stdout:

```
$MDRIP/go/bin/mdrip lesson1 \
    $MDRIP/go/src/github.com/monopole/mdrip/example_tutorial.md
```

## Tutorial Testing

The output of the above command can be piped to `source /dev/stdin` to
evolve the state of the current shell per the tutorial.

For automated testing it's better to pipe to `bash -e` (or some other
shell), running the code in a subshell leaving the current shell's
state unchanged (modulo whatever the script does to the computer).

Use of the tool's `--subshell` flag does that as well.  It assumes
bash blocks, but does a better job of reporting errors.
Run the following to see how the error in the example tutorial
is reported:

```
$MDRIP/go/bin/mdrip --subshell lesson1 \
    $MDRIP/go/src/github.com/monopole/mdrip/example_tutorial.md
```

The above command has no output and exits with status zero if all the
scripts labelled `@lesson1` in the given markdown succeed.  On any
failure, however, the command dumps a report and exits with non-zero
status.

## Alteratives

Another approach to testing or otherwise making code available to a user
shards the tutorial.

Code is placed in individual files in source control where it can be
tested via the usual mechanisms, then a server retrieves the code from
its canonical source and injects it into the markdown at serving time.
To make changes, one has to edit both the markdown file (the
discussion) and the various code files, plus maintain the injection
system.  A strictly worse approach uses a human as the injection
system - 'real' code (in some repository) and code in the discussion
(e.g.  on a wiki page) are distinct and have to be kept in sync
manually.

Nobody has time for either approach, so the code rots.
