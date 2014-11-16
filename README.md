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


## Example

Shell scripts can create files in any programming language, via
[_here_ documents](http://tldp.org/LDP/abs/html/here-docs.html) and
what not.

The following tutorial has shell snippets that write, compile and run
a Go program.

> The _markdown inside markdown_ problem skirted here by
> using single quotes to delimit the scripts rather than backticks.
> The 'real' version is [here](https://raw.githubusercontent.com/monopole/mdrip/master/example_tutorial.md),
> and the rendered version is [here](https://github.com/monopole/mdrip/blob/master/example_tutorial.md).

```
<!-- @1 @setup -->
'''
export GOPATH=/tmp/play/go
'''

Write a *Go* function:

<!-- @1 -->
'''
mkdir -p $GOPATH/src/example
 cat - <<EOF >$GOPATH/src/example/add.go
package main

func add(x, y int) (int) { return x + y }
EOF
'''

Write a main program to call it:

<!-- @1 -->
'''
 cat - <<EOF >$GOPATH/src/example/main.go
package main

import "fmt"

func main() {
    fmt.Printf("Calling add on 1 and 2 yields %d.\n", add(1, 2))
}
EOF
go install example
$GOPATH/bin/example
'''
Copy-paste the above to build and run your *Go* program.
```

The _@labels_ found in HTML comments directly preceeding the snippets
identify alternative 'threads' of execution.  The tool extracts
snippets with a given label.

The number of full scripts that can be generated equals the number of
unique labels.  If a snippet has multiple labels, it can be
incorporated into multiple scripts (e.g. common initialization code).
If a snippet has no label, it's ignored.

A simple script thread might merely put the user into an appropriate
shell state - proper environment vars defined, proper files created,
proper executables in place.

A more complex thread might build a server, run it in the background,
fire up a client to talk to it, then kill both through judicious used
of process ID variables.

## Build

Assuming `Go` and `git` are present:

```
export MDCHECK=/tmp/mdcheck
GOPATH=$MDCHECK/go go get github.com/monopole/mdrip
GOPATH=$MDCHECK/go go test github.com/monopole/mdrip
$MDCHECK/go/bin/mdrip   # Shows usage.
```

Extract code from the [example tutorial]
(https://github.com/monopole/mdrip/blob/master/example_tutorial.md)
with this command:

```
$MDCHECK/go/bin/mdrip \
    $MDCHECK/go/src/github.com/monopole/mdrip/example_tutorial.md 1
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
$MDCHECK/go/bin/mdrip --subshell \
    $MDCHECK/go/src/github.com/monopole/mdrip/example_tutorial.md 1
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
