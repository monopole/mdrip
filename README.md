This tool is a hacky, markdown-based instance of language-independent
[literate
programming](http://en.wikipedia.org/wiki/Literate_programming).  For
perspective, see the latex-based
[noweb](http://en.wikipedia.org/wiki/Noweb).

The tool scans markdown for [fenced code
blocks](https://help.github.com/articles/github-flavored-markdown/#fenced-code-blocks) immediately preceded by an HTML comment with an embeddded _@label_ and extracts them.

The tool can emit the code blocks to `stdout` for piping to an
arbitrary interpreter.  

If the code blocks are in bash syntax, and the tool is itself running
in a bash shell, then piping the output to `source /dev/stdin` is
equivalent to the user copying all the code blocks to their prompt.

Alternatively, the tool can run extracted code in a bash subshell with
customizable parsing of the subshell's stdout and stderr, allowing
reporting like _block 5 from script 'foo' in file 'bar' failed with
error 'baz'_.  This behavior facilitates adding tutorial coverage to
regression test frameworks.

## Details

The tool has a certainly rough extensibility because
shell scripts can make, build and run programs in any programming
language, via [_here_
documents](http://tldp.org/LDP/abs/html/here-docs.html) and what not.
The [example
tutorial](https://github.com/monopole/mdrip/blob/master/example_tutorial.md)
(raw markdown
[here](https://raw.githubusercontent.com/monopole/mdrip/master/example_tutorial.md)
) has _bash_ code blocks that write, compile and run a Go program.

The tool accepts a file argument and a label argument and extracts
all blocks with that label.

If a block has multiple labels, it can be incorporated into multiple
scripts.  If a block has no label, it's ignored.  The number
of scripts that can be extracted from a markdown file equals the
number of unique labels.

A block with a label like `@init` might merely define a few env
variables.  It might have a second label like `@lesson1` that also
appears on subsquent blocks that build a server, run it in the
background, fire up a client to talk to it, then kill both through
judicious used of process ID variables.  A final code block might
do cleanup, and have the label `@cleanup` so it can be run alone.

There's no notion of encapsulation or automatic cleanup.  Extracted 
scripts can do anything that the user can do.


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
$MDRIP/go/bin/mdrip \
    $MDRIP/go/src/github.com/monopole/mdrip/example_tutorial.md lesson1
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
$MDRIP/go/bin/mdrip --subshell \
    $MDRIP/go/src/github.com/monopole/mdrip/example_tutorial.md lesson1
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
