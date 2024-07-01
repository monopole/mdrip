package monkey

import _ "embed"

var (
	//go:embed monkey.js
	Js string
)
