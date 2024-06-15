# The gomarkdown parser

https://github.com/gomarkdown/markdown

Notes made in late 2023

## GOOD
 - It has no open pull requests (responsive owners)
 - Better documentation than goldmark.
 - Clear access to the AST, as the API requires you to hold it 
   in between
 - The AST has all the document contents.
 - It supports block level attributes: {#id3 .myclass fontsize="tiny"}' on (at least)
   header blocks and code blocks
 # Meh
 - It could support mermaid : https://github.com/gomarkdown/markdown/issues/284,
   but it's not clear if anyone did so.
 - The number of contributors is unclear, since it's a fork of blackfriday.
 - It has zero official releases.
