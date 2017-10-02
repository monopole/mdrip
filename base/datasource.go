package base

import (
	"errors"
	"strings"
)

type DataSource struct {
	args []string
}

func (d *DataSource) FirstArg() string {
	return d.args[0]
}

func (d *DataSource) N() int {
	return len(d.args)
}

func (d *DataSource) AsPaths() []FilePath {
	result := make([]FilePath, len(d.args))
	for i, x := range d.args {
		result[i] = FilePath(x)
	}
	return result
}

func (d *DataSource) String() string {
	n := d.args[0]
	if len(d.args) > 1 {
		n += "..."
	}
	return n
}

func NewDataSource(fArgs []string) (*DataSource, error) {
	result := []string{}
	for _, n := range fArgs {
		n := strings.TrimSpace(n)
		if len(n) > 0 {
			result = append(result, n)
		}
	}
	if len(result) < 1 {
		return nil, errors.New("Must specify a data source - files, directory, or github clone url.")
	}
	return &DataSource{result}, nil
}
