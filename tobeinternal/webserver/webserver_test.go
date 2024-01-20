package webserver

import (
	"testing"

	"github.com/monopole/mdrip/tobeinternal/base"
	"github.com/monopole/mdrip/tobeinternal/loaderold"
)

func TestNewWebServer(t *testing.T) {
	ds, err := base.NewDataSet([]string{"hey"})
	if err != nil {
		t.Errorf("trouble with datasource")
		return
	}
	l := loaderold.NewLoader(ds)
	_, err = NewServer(l)
	if err != nil {
		t.Errorf("unable to make server: %v", err)
		return
	}
}
