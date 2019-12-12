module github.com/monopole/mdrip

go 1.13

require (
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/gorilla/context v0.0.0-20160226214623-1ea25387ff6f // indirect
	github.com/gorilla/mux v1.6.0
	github.com/gorilla/securecookie v0.0.0-20160422134519-667fe4e3466a // indirect
	github.com/gorilla/sessions v0.0.0-20160922145804-ca9ada445741
	github.com/gorilla/websocket v1.2.0
	github.com/pkg/errors v0.8.0
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	gopkg.in/russross/blackfriday.v2 v2.0.0
)

// v2.0.0 is incompatible with kubectl libraries
exclude github.com/russross/blackfriday v2.0.0+incompatible
