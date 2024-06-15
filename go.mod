module github.com/monopole/mdrip/v2

go 1.21

require (
	github.com/gomarkdown/markdown v0.0.0-20231115200524-a660076da3fd
	github.com/gorilla/mux v1.8.1
	github.com/gorilla/sessions v1.2.2
	github.com/gorilla/websocket v1.5.1
	github.com/monopole/shexec v0.1.8
	github.com/spf13/afero v1.11.0
	github.com/spf13/cobra v1.8.0
	github.com/stretchr/testify v1.8.4
	github.com/tdewolff/minify/v2 v2.20.31
	github.com/yosssi/gohtml v0.0.0-20201013000340-ee4748c638f4
	github.com/yuin/goldmark v1.6.0

)

//replace (
//	github.com/gomarkdown/markdown => ../../gomarkdown/markdown
//	github.com/monopole/shexec => ../shexec
//	github.com/yuin/goldmark => ../../yuin/goldmark
//)

exclude github.com/monopole/mdrip v1.0.3

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gorilla/securecookie v1.1.2 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/tdewolff/parse/v2 v2.7.14 // indirect
	golang.org/x/net v0.19.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
