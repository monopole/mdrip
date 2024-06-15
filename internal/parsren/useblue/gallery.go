package useblue

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/gomarkdown/markdown/ast"
)

type Gallery struct {
	ast.Leaf
	ImageURLS []string
}

var _ ast.Node = &Gallery{}

var gallery = []byte(":gallery\n")

func attemptToParseGallery(data []byte) (*Gallery, []byte, int) {
	if !bytes.HasPrefix(data, gallery) {
		return nil, nil, 0
	}
	fmt.Printf("Found a gallery!\n\n")
	i := len(gallery)
	// find empty line
	// TODO: should also consider end of document
	end := bytes.Index(data[i:], []byte("\n\n"))
	if end < 0 {
		return nil, data, 0
	}
	end = end + i
	return &Gallery{
		ImageURLS: strings.Split(string(data[i:end]), "\n"),
	}, nil, end
}
