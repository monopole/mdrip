package base

import (
	"errors"
)

type DataSet struct {
	args []*DataSource
}

func (d *DataSet) FirstArg() *DataSource {
	return d.args[0]
}

func (d *DataSet) N() int {
	return len(d.args)
}

func (d *DataSet) AsPaths() []FilePath {
	result := make([]FilePath, len(d.args))
	for i, x := range d.args {
		result[i] = x.AbsPath()
	}
	return result
}

func (d *DataSet) String() string {
	n := d.args[0].Display()
	if len(d.args) > 1 {
		n += "..."
	}
	return n
}

func NewDataSet(fArgs []string) (*DataSet, error) {
	result := []*DataSource{}
	for _, n := range fArgs {
		item, err := NewDataSource(n)
		if err != nil {
			return nil, errors.New("Bad arg " + n)
		}
		result = append(result, item)
	}
	if len(result) < 1 {
		return nil, errors.New("Must specify a data source - files, directory, or github clone url.")
	}
	return &DataSet{result}, nil
}
