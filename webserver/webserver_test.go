package webserver

import (
"testing"
	"github.com/monopole/mdrip/base"
	"github.com/monopole/mdrip/loader"
)

func TestNewWebServer(t *testing.T) {
	ds, err := base.NewDataSource([]string{"hey"})
	if err != nil {
		t.Errorf("trouble with datasource")
		return
	}
	l := loader.NewLoader(ds)
	_, err = NewServer(l)
	if err != nil {
		t.Errorf("unable to make server: %v", err)
		return
	}
}
