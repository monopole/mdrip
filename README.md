# mdrip

[literate programming]: http://en.wikipedia.org/wiki/Literate_programming
[_here_ documents]: http://tldp.org/LDP/abs/html/here-docs.html
[busted Go tutorial]: assets/bustedGoTutorial.md
[raw]: https://github.com/monopole/mdrip/blob/master/assets/bustedGoTutorial.md?plain=1
[travis-mdrip]: https://travis-ci.org/monopole/mdrip
[`tmux`]: https://github.com/tmux/tmux/wiki
[fenced code blocks]: https://help.github.com/articles/creating-and-highlighting-code-blocks/#fenced-code-blocks
[block quote]: https://github.github.com/gfm/#block-quotes
[label]: #labels
[labels]: #labels
[release page]: https://github.com/monopole/mdrip/releases

<!-- [![Build Status](https://travis-ci.org/monopole/mdrip.svg?branch=master)](https://travis-ci.org/monopole/mdrip) -->
[![Go Report Card](https://goreportcard.com/badge/github.com/monopole/mdrip/v2)](https://goreportcard.com/report/github.com/monopole/mdrip/v2)

`mdrip` is a markdown code block extractor.

To extract and print all code blocks below your current directory:

> ```
> mdrip print .
> ```

Pipe that into `/bin/bash -e` to have the effect of a test, or better yet
try the `test` command:

> ```
> mdrip test .
> ```

This fails only if an extracted markdown code block fails.
Use a [label] to be selective about which blocks to run.

To give demos with [`tmux`], or generally use your browser and `tmux`
as a markdown code block IDE, try:

> ```
> mdrip serve --port 8080 .
> ```

In the browser, while focused on a code block, hit the `Enter` key.
The app posts the block's ID to the server, and the server sends
the corresponding code block to `tmux` via its api.

<a href="assets/mdripDemo.png" target="_blank">
<img src="assets/mdripDemo.png"
  alt="mdrip screenshot" width="95%" height="auto">
</a>


## Installation

Install via the [Go](https://golang.org/dl) tool:
<!-- @installation -->
```
go install github.com/monopole/mdrip/v2@latest
```
or download a build from the [release page].

## Basic Extraction and Testing

For something to work with,
download this [busted Go tutorial]:

<!-- @downloadBusted -->
```
repo=https://raw.githubusercontent.com/monopole/mdrip
curl -O $repo/master/assets/bustedGoTutorial.md
```

This markdown has code blocks showing how to write, compile
and run a Go program in your `TMPDIR`.

Extract the blocks to `stdout`:

<!-- @lookAtBlocks -->
```
mdrip print bustedGoTutorial.md
```

Some code blocks in this markdown have [labels]; these are visible as
HTML comments preceding the blocks in the [raw] markdown.

Use a label to extract a subset of blocks:
<!-- @useLabel -->
```
mdrip print --label goCommand bustedGoTutorial.md
```

Test the code from the markdown in a subshell:

<!-- @testTheBlocks -->
```
mdrip test bustedGoTutorial.md
echo $?
```

The above command should show an error, and exit with non-zero status,
because the tutorial has errors.

Fix the error:

<!-- @copyTheTutorial -->
```
cp bustedGoTutorial.md goTutorial.md
```

<!-- @fixError1 -->
```
sed -i 's|badecho |echo |' goTutorial.md
```

Try the fix:

<!-- @tryFix1 -->
```
mdrip test goTutorial.md
echo $?
```

There's another error.  Fix it:

<!-- @fixError2 -->
```
sed -i 's|comment this|// comment this|' goTutorial.md
```

There are now two changes:

<!-- @observeDiffs -->
```
diff bustedGoTutorial.md goTutorial.md
```

Test the new file:

<!-- @testAgain -->
```
mdrip test goTutorial.md
echo $?
```

The return code should be zero.

You can run a block in your _current_ shell to, say, set
current environment variables as specified in the markdown:

<!-- @evalInShell -->
```
eval "$(mdrip print --label setEnv goTutorial.md)"
echo $greeting
```

The upshot is that adding a line like

> ```
> mdrip test --label {someLabel} {filePath}
> ```

to your CI/CD test framework covers
the markdown code block execution path determined by that label.


The `{filepath}` argument defaults to your current directory (`.`),
but it can be 

* a path to a file,
* a path to a directory,
* a GitHub URL in the style `gh:{user}/{repoName}`,
* or a particular file or a directory in the
  repo, e.g. `gh:{user}/{repoName}/foo/bar`.

### Labels

Add _labels_ to a code block by preceding the block
with a one-line HTML comment, e.g:

<blockquote>
<pre>
&lt;&#33;-- @sayHello @mississippi @tutorial01 --&gt;
&#96;&#96;&#96;
echo hello
&#96;&#96;&#96;
</pre>
</blockquote>

Labels are just words beginning with `@` in the comment.

The first label on a block is slightly special in that it
is treated as the block's _name_ for various purposes.
If no labels are present, a block name is generated for these
purposes.

## Demonstrations and Tutorial Development

`mdrip` and [`tmux`] provide a handy way to develop and demonstrate
command line procedures.

Render a markdown web app like this:
<!-- @serveTutorial -->
```
mdrip serve --port 8000 goTutorial.md
```
Visit it at [localhost:8000]([http://localhost:8000].).

Hit the `n` key for navigation tools.
Hit `?` to see all key controls.

The handy aspect is provided by [`tmux`].
If there's a running instance of `tmux`, the server
will send code blocks to it when you hit `Enter`.

Fire up `tmux`, then try this `README` directly:

<!-- @serveMdripReadme -->
```
mdrip serve gh:monopole/mdrip/README.md
```

To see what using a full tree of markdown looks like,
generate some content with:
<!-- @createTestData -->
```
mdrip writemd /tmp/mdTestData
```
then serve it:
<!-- @serveTestData -->
```
mdrip serve /tmp/mdTestData
```


## Tips

`mdrip` encourages [literate programming] via markdown.

It lets one run or test code (shell commands) that is otherwise
embedded in explanatory content (markdown).

One can use [_here_ documents] to incorporate _any_ programming
language into tested markdown - as in the [busted Go tutorial]
discussed above.  That tutorial could have covered C, C++, Rust, etc.

Place commands that the reader would want to execute directly
(with no edits) in [fenced code blocks].

In contrast, code-style text that is not intended for copy/paste execution,
e.g. alternative commands with fake arguments or example output,
should be in a fenced code block indented via a
[block quote]. Block quotes are ignored by `mdrip`.

Eschew adding prompts to code blocks.
The following code snippet is easy to copy/paste:
> ```
> echo hello
> du -sk
> ```
But this is not:
> ```
> $ echo hello
> $ du -sk
> ```
