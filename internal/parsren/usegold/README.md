# The goldmark parser

https://github.com/yuin/goldmark

Notes made in _Fall 2023_

## GOOD
 - One active, dedicated, awesome maintainer.
 - lots of extensions, proven framework.
 - goldmark is now the markdown renderer for Hugo, replacing blackfriday
 - It already supports mermaid via an extension.
 - It has (at least) 80 releases!  https://github.com/yuin/goldmark/releases
 - (at least) 80% coverage

### MEH
 - Some PRs being ignored by the maintainer. Sure.
 - It doesn't yet support block level attributes, but maybe it's coming.
 - The parser API doesn't seem to return errors. Odd.


Since this package renders HTML, the  `webapp` depends on it,
and this dependence seeps into various things.
