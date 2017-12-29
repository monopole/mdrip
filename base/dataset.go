package base

import (
	"errors"
)

// DataSet indicates the origin of multiple markdown sources.
type DataSet struct {
	args []*DataSource
}

// FirstArg is, uh, the first member of the dataset - sometimes special.
func (d *DataSet) FirstArg() *DataSource {
	return d.args[0]
}

// Size is the dataset size.
func (d *DataSet) Size() int {
	return len(d.args)
}

// AsPaths is an array of file paths representing the dataset.
func (d *DataSet) AsPaths() []FilePath {
	result := make([]FilePath, len(d.args))
	for i, x := range d.args {
		result[i] = x.AbsPath()
	}
	return result
}

// String is a string form for the dataset.
func (d *DataSet) String() string {
	n := d.args[0].Display()
	if len(d.args) > 1 {
		n += "..."
	}
	return n
}

// NewDataSet makes a dataset from the given args.
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
		return nil, errors.New("must specify a data source - files, directory, or github clone url")
	}
	return &DataSet{result}, nil
}
