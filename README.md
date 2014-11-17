# Literate programming for markdown

This tool reads markdown and extracts labelled code blocks.

The tool can emit the code to `stdout` for piping to `source
/dev/stdin`, as if the user had copy/pasted the blocks to their
prompt.  This is a way to fast-forward through all or part of
said code to get into a desired state.

Alternatively, the tool can run extracted code in a subshell with
customizable parsing of the subshell's stdout and stderr, allowing
reporting like _subscript 5 from thread 'foo' in file 'bar' failed
with error 'baz'_.  This behavior facilitates adding tutorial
coverage to regression test frameworks.

This tool is a hacky, markdown-based instance of language-independent
[literate
programming](http://en.wikipedia.org/wiki/Literate_programming).  For
perspective, see the latex-based
[noweb](http://en.wikipedia.org/wiki/Noweb).


## Details

This tool looks for [fenced code
blocks](https://help.github.com/articles/github-flavored-markdown/#fenced-code-blocks).
The blocks are viewed as shell scripts, and shell scripts can make
files in any programming language, via [_here_
documents](http://tldp.org/LDP/abs/html/here-docs.html) and what not.

The tutorial
[here](https://github.com/monopole/mdrip/blob/master/example_tutorial.md)
(raw markdown
[here](https://raw.githubusercontent.com/monopole/mdrip/master/example_tutorial.md)
) has code blocks that write, compile and run a Go program.

An _@labels_ found in a _HTML comment_ immediately preceeding a code
block identify the block as part of a _labelled script_.  

The tool accepts a label argument and file argument and extracts the
matching script.

If a block has multiple labels, it can be incorporated into multiple
scripts (e.g. common initialization code).  If a block has no label,
this tool ignores it.  The number of scripts that can be extracted
from a markdown file equals the number of unique labels.

A block with a label like `@init` might merely define a few env
variables.  It might have a second label like `@lesson1` that also
appears on subsquent blocks that build a server, run it in the
background, fire up a client to talk to it, then kill both through
judicious used of process ID variables.

There's no notion of encapsulation or automatic cleanup.  Extracted 
scripts can do anything that the user can do.

The markdown author can always add cleanup code to the final block.

## Build

Assuming `Go` and `git` are present:

```
export MDRIP=~/mdrip
GOPATH=$MDRIP/go go get github.com/monopole/mdrip
GOPATH=$MDRIP/go go test github.com/monopole/mdrip
$MDRIP/go/bin/mdrip   # Shows usage.
```

Extract code from the [example tutorial]
(https://github.com/monopole/mdrip/blob/master/example_tutorial.md)
with this command:

```
$MDRIP/go/bin/mdrip \
    $MDRIP/go/src/github.com/monopole/mdrip/example_tutorial.md 1
```

## Testing

The output of the above command can be piped to `source /dev/stdin` to
evolve the state of the current shell per the tutorial.

For automated testing, however, it's better to pipe to `bash -e`,
running the code in a subshell leaving the current shell's state
unchanged (modulo whatever the script does to the computer).

Use of the `--subshell` flag does that as well - but does a better job
of reporting errors:

```
$MDRIP/go/bin/mdrip --subshell \
    $MDRIP/go/src/github.com/monopole/mdrip/example_tutorial.md 1
```

The above command has no output and exits with status zero if all the
scripts labelled `@1` in the given markdown succeed.  On any failure,
however, the command dumps a report and exits with non-zero status.

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
