# The goldmark parser

https://github.com/yuin/goldmark

Notes made in late 2023; likely very out of date

## GOOD
 - One active, dedicated, awesome maintainer.
 - lots of extensions, proven framework.
 - goldmark is now the markdown renderer for Hugo, replacing blackfriday
 - It already supports mermaid via an extension.
 - It has (at least) 80 releases!  https://github.com/yuin/goldmark/releases
 - (at least) 80% coverage

### maybe problems
 - Some PRs being ignored by the maintainer. Sure.
 - It doesn't yet support block level attributes, but maybe it's coming.
 - The parser API doesn't seem to return errors.
 - The rendering aspect of the package is pretty tightly bound 
   to its parsing aspect.
 
