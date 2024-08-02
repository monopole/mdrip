#!/bin/bash

# Build the current workspace, injecting loader flag values
# so that the version command works.

ldPath=github.com/monopole/mdrip/v2/internal/provenance

# Use a date format compatible with
# https://pkg.go.dev/time#pkg-constants RFC3339
buildDate=$(date -u +'%Y-%m-%dT%H:%M:%SZ')

version=$(git describe --tags --always --dirty)

# Assume that the code is modified.
gitCommit="$(git branch --show-current)-modified"

out=$(go env GOPATH)/bin/mdrip
/bin/rm -f $out

go build \
  -o $out \
	-ldflags \
  "-X ${ldPath}.version=${version} \
	 -X ${ldPath}.gitCommit=${gitCommit} \
	 -X ${ldPath}.buildDate=${buildDate}" \
  .
