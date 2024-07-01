# widgets

Widgets are file triplets of `widget.js`,
`widget.html` and `widget.css`. The latter two
aren't needed if the widget is just a js library.
The `.go` files hold glue code to ease rendering.

The files with `.js`, `.html` and `.css` file name
extensions are actually Go templates - in that
they have moustache variables, etc.  But they
are *mostly* javascript, HTML and CSS, and using
these file extensions eases editing them in an IDE.

There should be no global variables  in the `.js`
files; they should hold only class declarations.

The `widget.html` documents hold only an HTML
snippet, typically bounded by `<div> ... </div>`.
These files are not intended to be complete
HTML documents with `<head>` and `<body>` tags.

Full document tags are provided by an
encapsulating test or web app.

