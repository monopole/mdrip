package provenance_test

import (
	"fmt"
	"runtime/debug"
	"testing"

	. "github.com/monopole/mdrip/v2/internal/provenance"
	"github.com/stretchr/testify/assert"
)

func TestGetProvenance(t *testing.T) {
	p := GetProvenance()
	assert.Equal(t, DefaultVersion, p.Version)
	assert.Equal(t, DefaultBuildDate, p.BuildDate)
	// This comes from BuildInfo,
	// which is not set during go test: https://github.com/golang/go/issues/33976
	assert.Equal(t, DefaultGitCommit, p.GitCommit)

	// These are set properly during go test
	assert.NotEmpty(t, p.GoArch)
	assert.NotEmpty(t, p.GoOs)
	assert.Contains(t, p.GoVersion, "go1.")
}

func TestProvenance_Short(t *testing.T) {
	p := GetProvenance()
	assert.Equal(
		t, fmt.Sprintf("{%s  %s   }", DefaultVersion, DefaultBuildDate), p.Short())
}

func mockModule(version string) debug.Module {
	return debug.Module{
		Path:    "github.com/monopole/mdrip/v2",
		Version: version,
		Replace: nil,
	}
}

func TestGetMostRecentTag(t *testing.T) {
	tests := []struct {
		name        string
		module      debug.Module
		isError     bool
		expectedTag string
	}{
		{
			name:        "Standard version",
			module:      mockModule("v1.2.3"),
			expectedTag: "v1.2.3",
		},
		{
			name:        "Pseudo-version with patch increment",
			module:      mockModule("v0.0.0-20210101010101-abcdefabcdef"),
			expectedTag: "v0.0.0",
		},
		{
			name:    "Invalid semver string",
			module:  mockModule("invalid-version"),
			isError: true,
		},
		{
			name:        "Valid semver with patch increment and pre-release info",
			module:      mockModule("v1.2.3-0.20210101010101-abcdefabcdef"),
			expectedTag: "v1.2.2",
		},
		{
			name:        "Valid semver no patch increment",
			module:      mockModule("v1.2.0"),
			expectedTag: "v1.2.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag, err := GetMostRecentTag(tt.module)
			if err != nil {
				if !tt.isError {
					assert.NoError(t, err)
				}
			} else {
				assert.Equal(t, tt.expectedTag, tag)
			}
		})
	}
}
