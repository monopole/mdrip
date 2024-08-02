# mdrip

[literate programming]: http://en.wikipedia.org/wiki/Literate_programming
[_here_ documents]: http://tldp.org/LDP/abs/html/here-docs.html
[busted Go tutorial]: assets/bustedGoTutorial.md
[raw]: https://raw.githubusercontent.com/monopole/mdrip/master/hack/bustedGoTutorial.md
[travis-mdrip]: https://travis-ci.org/monopole/mdrip
[`tmux`]: https://github.com/tmux/tmux/wiki
[fenced code blocks]: https://help.github.com/articles/creating-and-highlighting-code-blocks/#fenced-code-blocks
[block quote]: https://github.github.com/gfm/#block-quotes
[label]: #labels


<!-- [![Build Status](https://travis-ci.org/monopole/mdrip.svg?branch=master)](https://travis-ci.org/monopole/mdrip) -->
[![Go Report Card](https://goreportcard.com/badge/github.com/monopole/mdrip/v2)](https://goreportcard.com/report/github.com/monopole/mdrip/v2)

Extract and run all markdown code blocks in your current directory:
> ```
> mdrip print . | /bin/bash -e
> ```

This exits with the error code of the first failed command.
Use a [label] to be selective about which blocks to run.

Convert your markdown into a web app that works with [`tmux`]:

> ```
> mdrip serve --port 8080 .
> ```

While focused on a code block, hit the `Enter` key.
The app posts the block's ID to the server, and the server sends
the corresponding code block to `tmux` via its api.

<a href="assets/mdripDemo.png" target="_blank">
<img src="assets/mdripDemo.png"
  alt="mdrip screenshot" width="95%" height="auto">
</a>


## Installation

Visit the [release page](https://github.com/monopole/mdrip/releases)
and download the executable for your OS and architecture.

If you have [Go](https://golang.org/dl) installed you can try:

<!-- @installation -->
```
go install github.com/monopole/mdrip/v2@latest
```

## Testing

For something to work with,
download the [busted Go tutorial] ([raw]):

<!-- @downloadBusted -->
```
repo=https://raw.githubusercontent.com/monopole/mdrip
curl -O $repo/master/hack/bustedGoTutorial.md
```

It has code blocks that seek to write, compile
and run a Go program in your `TMPDIR`.

Extract blocks to `stdout`:

<!-- @lookAtBlocks -->
```
echo " "
mdrip print bustedGoTutorial.md
```

Extract a subset of blocks by using a [label]:
<!-- @useLabel -->
```
echo " "
mdrip print --label goCommand bustedGoTutorial.md
```

Test the code from the markdown in a subshell:

<!-- @pipeTheBlocks -->
```
echo " "
mdrip print bustedGoTutorial.md | bash -e
echo $?
```

The above command should show an error, and exit with non-zero status,
because the tutorial has errors.

For a clearer description of the error, try the `test` command:

<!-- @testTheBlocks -->
```
echo " "
mdrip test bustedGoTutorial.md
echo $?
```
This also fails.

Fix the errors:

<!-- @copyTheTutorial -->
```
cp bustedGoTutorial.md goTutorial.md
```

<!-- @fixTutorial -->
```
sed -i 's|comment this|// comment this|' goTutorial.md
sed -i 's|badecho |echo |' goTutorial.md
```

<!-- @observeDiffs -->
```
diff bustedGoTutorial.md goTutorial.md
```

Test the new file:

<!-- @pipeAgain -->
```
echo " "
mdrip print goTutorial.md | bash -e
echo $?
```

or to get quieter output:
<!-- @testAgain -->
```
echo " "
mdrip test goTutorial.md
echo $?
```

The return code should be zero.

Run a block in your _current_ shell to, say, set
current environment variables as specified in the markdown:

<!-- @evalInShell -->
```
eval "$(mdrip print --label setEnv goTutorial.md)"
echo $greeting
```

The upshot is that adding a line like

> ```
> mdrip print --label {someLabel} {filePath} | /bin/bash -e
> ```

to your CI/CD test framework covers
the execution path determined by that label.


The `{filepath}` argument can be

* a single local file,
* a local directory,
* a github URL in the style `gh:{user}/{repoName}`,
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

The first label on a block is slightly special in
that it is reported as the block's name for various
purposes. If no labels are present, a unique label
will be generated for these purposes.

## Demonstrations

`mdrip` is a handy way to demonstrate command line procedures.

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

To see what using a full tree of markdown looks like, generate
some content with:
<!-- @createTestData -->
```
mdrip gentestdata /tmp/mdTestData
```
then serve it:
<!-- @serveTestData -->
```
mdrip serve /tmp/mdTestData
```



## Tips

`mdrip` permits [literate programming] via markdown.
It lets one run or test code (shell commands) that is otherwise
embedded in explanatory content (markdown).
One can use [_here_ documents] to incorporate any programming
language into tested markdown - as in the [busted Go tutorial]
discussed above.  That tutorial could have covered C, C++, Rust, etc.

Place commands that the reader would want to execute directly
(with no edits) in [fenced code blocks].

In contrast, code-style text that is _not_ intended for copy/paste execution,
e.g. alternative commands with fake arguments or example output,
should be in a fenced code block indented via a
[block quote]. Block quotes are ignored by `mdrip`.

Eschew adding prompts to code blocks.
The following code snippet is easy to copy/paste:
> ```
> echo hello
> du -sk
> ```
But this isn't:
> ```
> $ echo hello
> $ du -sk
> ```
