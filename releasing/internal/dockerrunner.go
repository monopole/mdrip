package internal

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DockerRunner runs some "docker" commands.
type DockerRunner struct {
	rn      *MyRunner
	ldVars  *LdVars
	pgmName string
	dirTmp  string
}

const (
	imageRegistry = "hub.docker.com"
	imageOwner    = "monopole"

	dockerTemplate = `
# This file is generated; DO NOT EDIT.
FROM golang:1.22.5-bullseye
WORKDIR /go/src/github.com/monopole/{{PGMNAME}}
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOWORK=off \
  go build -v -o /go/bin/{{PGMNAME}} \
  -ldflags "{{LDFLAGS}}" \
  .
ENTRYPOINT ["/go/bin/{{PGMNAME}}"]
`
)

func NewDockerRunner(dirSrc, dirTmp string, ldVars *LdVars) *DockerRunner {
	return &DockerRunner{
		rn:      NewMyRunner("docker", dirSrc, DoIt, 3*time.Minute),
		ldVars:  ldVars,
		dirTmp:  dirTmp,
		pgmName: filepath.Base(dirSrc),
	}
}

func (dr *DockerRunner) Content() []byte {
	content := strings.Replace(
		dockerTemplate[1:], "{{LDFLAGS}}", dr.ldVars.MakeLdFlags(), -1)
	content = strings.Replace(content, "{{PGMNAME}}", dr.pgmName, -1)
	return []byte(content)
}

func (dr *DockerRunner) ImageName() string {
	return imageOwner + "/" + dr.pgmName + ":" + dr.ldVars.Version()
}

func (dr *DockerRunner) Build() error {
	dr.rn.comment("building docker image at tag " + dr.ldVars.Version())
	dockerFileName := filepath.Join(dr.dirTmp, "Dockerfile")
	if err := os.WriteFile(dockerFileName, dr.Content(), 0644); err != nil {
		return err
	}
	slog.Info("Wrote", "file", dockerFileName)
	err := dr.rn.run(
		NoHarmDone,
		"build",
		"--file", dockerFileName,
		// "--platform", "linux/amd64,linux/arm64" (not using this yet)
		"-t", dr.ImageName(),
		".",
	)
	if err != nil {
		slog.Error(err.Error())
		slog.Error(dr.rn.Out())
		return err
	}
	return nil
}

func (dr *DockerRunner) Push() error {
	err := dr.rn.run(
		UndoIsHard,
		"push",
		dr.ImageName(),
	)
	if err != nil {
		slog.Error(err.Error())
		slog.Error(dr.rn.Out())
	}
	return err
}

func (dr *DockerRunner) Login() error {
	dur := dr.rn.duration
	dr.rn.duration = 3 * time.Second
	err := dr.rn.run(
		NoHarmDone,
		"login",
	)
	dr.rn.duration = dur
	if err != nil {
		slog.Error(err.Error())
		slog.Error(dr.rn.Out())
	}
	return err
}
