# Parsers and renderers

In the beginning, `mdrip` had its own custom
markdown parser and its own custom markdown renderer.

As fun as that was, time passed and folks created many extensions
to markdown; tables of contents, mermaid diagrams, etc.

Both parsing and rendering have gotten better dedicated parsers
and renders with plugin architectures, exceeding the capabilities
of the parser and renderer builtin to `mdrip`.

So the original parser and renderer was retired, to be replaced
by one or both of these:

  * https://github.com/yuin/goldmark
  * https://github.com/gomarkdown/markdown

The  `MdParserRenderer` interface was defined - an `mdrip`-specific
interface.  The idea was to slap an adapter over each of these two
packages allowing them to implement this interface.  Then provide
an `mdrip` command line flag to pick one of the two implementations
at run time and inject it.

Work began on adapting `gomarkdown/markdown` to the interface,
but paused when it became clear that `yuin/goldmark` was working fine.
Keeping this arrangement for a while, just in case it becomes
desirable to try `gomarkdown/markdown` again.


