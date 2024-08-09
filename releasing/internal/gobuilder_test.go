package internal

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLdVars_MakeLdFlags(t *testing.T) {
	tests := map[string]struct {
		ImportPath string
		Kvs        map[string]string
		want       string
	}{
		"t1": {
			ImportPath: "github.com/foo/bar/provenance",
			Kvs:        map[string]string{"fruit": "apple", "animal": "dog"},
			want: `-s -w ` +
				`-X github.com/foo/bar/provenance.fruit=apple ` +
				`-X github.com/foo/bar/provenance.animal=dog`,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ldv := &LdVars{
				ImportPath: tc.ImportPath,
				Kvs:        tc.Kvs,
			}
			assert.Equalf(t, tc.want, ldv.MakeLdFlags(), "MakeLdFlags()")
		})
	}
}
