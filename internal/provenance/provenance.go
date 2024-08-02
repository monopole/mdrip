package provenance

import (
	"fmt"
	"github.com/blang/semver"
	"runtime"
	"runtime/debug"
	"strings"
)

const (
	// DefaultVersion should match the value debug.BuildInfo uses for an unset
	// main module version.
	DefaultVersion   = "(devel)"
	DefaultGitCommit = "gcUnknown"
	DefaultBuildDate = "bdUnknown"
)

// These variables may be overwritten at build time using ldflags.
//
//nolint:gochecknoglobals
var (
	version   = DefaultVersion
	gitCommit = DefaultGitCommit
	// build date in time.RFC3339 format
	buildDate = DefaultBuildDate
)

// Provenance holds information about the build of an executable.
type Provenance struct {
	// Version of the binary, assumed to be in semver format.
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
	// GitCommit is a git commit
	GitCommit string `json:"gitCommit,omitempty" yaml:"gitCommit,omitempty"`
	// BuildDate is date of the build.
	BuildDate string `json:"buildDate,omitempty" yaml:"buildDate,omitempty"`
	// GoOs holds OS name.
	GoOs string `json:"goOs,omitempty" yaml:"goOs,omitempty"`
	// GoArch holds architecture name.
	GoArch string `json:"goArch,omitempty" yaml:"goArch,omitempty"`
	// GoVersion holds Go version.
	GoVersion string `json:"goVersion,omitempty" yaml:"goVersion,omitempty"`
}

const experiment = false

// GetProvenance returns an instance of Provenance.
func GetProvenance() Provenance {
	p := Provenance{
		BuildDate: buildDate,
		Version:   version,
		GitCommit: gitCommit,
		GoOs:      runtime.GOOS,
		GoArch:    runtime.GOARCH,
		GoVersion: runtime.Version(),
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return p
	}
	for _, setting := range info.Settings {
		// For now, the git commit is the only information of interest.
		// We could consider adding other info such as the commit date
		// in the future.
		if experiment && setting.Key == "vcs.revision" {
			//	p.GitCommit = setting.Value
		}
	}
	if !experiment {
		return p
	}
	for _, dep := range info.Deps {
		if dep != nil && dep.Path == "github.com/monopole/mdrip/v2" {
			if dep.Version != "devel" {
				continue
			}
			v, err := GetMostRecentTag(*dep)
			if err != nil {
				fmt.Printf(
					"failed to get most recent tag for %s: %v\n", dep.Path, err)
				continue
			}
			p.Version = v
		}
	}
	return p
}

func GetMostRecentTag(m debug.Module) (string, error) {
	for m.Replace != nil {
		m = *m.Replace
	}

	split := strings.Split(m.Version, "-")
	sv, err := semver.Parse(strings.TrimPrefix(split[0], "v"))

	if err != nil {
		return "", fmt.Errorf("failed to parse version %s: %w", m.Version, err)
	}

	if len(split) > 1 && sv.Patch > 0 {
		sv.Patch -= 1
	}
	return "v" + sv.String(), nil
}

// Short returns the shortened provenance stamp.
func (v Provenance) Short() string {
	return fmt.Sprintf(
		"%v",
		Provenance{
			Version:   v.Version,
			BuildDate: v.BuildDate,
		})
}
