#!/bin/bash
# Copyright 2023 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

#
# This script is called by Kustomize's release pipeline.
# It needs jq (required for release note construction) and [GitHub CLI](https://cli.github.com/).
#
# To test it locally:
#
#   # Please install jq and GitHub CLI. (e.g. macOS)
#   brew install jq gh
#
#   # Setup GitHub CLI
#   gh auth login
#
#   # Run this script, where $TAG is the tag to "release" (e.g. kyaml/v0.13.4)
#   ./releasing/create-release.sh $TAG
#
#   # Please remove Draft Release created by this script.

set -o errexit
set -o nounset
set -o pipefail

#if [[ -z "${1-}" ]]; then
#  echo "Usage: $0 {tag}"
#  echo "  {tag}: the tag to build or release, e.g. v1.2.3"
#  exit 1
#fi

# gitTag=$1
# echo "release tag: $gitTag"

# Build release binaries for every OS/arch combination, place in $releaseDir.
function buildBinaries {
  local version=$1
  local releaseDir=$2
  local ldVarPath="github.com/monopole/mdrip/internal"
  # build date in ISO8601 format
  local buildDate=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
  local osList=(linux darwin windows)
  local workDir="tmpBuildOutput"

  mkdir -p $workDir

  for os in "${osList[@]}"; do
    archList=(amd64 arm64)
    # if [ "$os" == "linux" ]; then
    #   archList=(amd64 arm64 s390x ppc64le)
    # fi
    for arch in "${archList[@]}" ; do
      echo "Building $os-$arch"
      binaryName="mdrip"
      [[ "$os" == "windows" ]] && binaryName="mdrip.exe"
      CGO_ENABLED=0 GOOS=$os GOARCH=$arch \
        go build -o output/$binaryName -ldflags\
        "-s -w\
         -X ${ldVarPath}/provenance.version=$version\
         -X ${ldVarPath}/provenance.gitCommit=$(git rev-parse HEAD)\
         -X ${ldVarPath}/provenance.buildDate=$buildDate"\
        main.go
      artifact="${releaseDir}/mdrip_${version}_${os}_${arch}"
      if [ "$os" == "windows" ]; then
        zip -j ${artifact}.zip ${workDir}/$binaryName
      else
        tar cvfz ${artifact}.tar.gz -C ${workDir} $binaryName
      fi
      rm ${workDir}/$binaryName
    done
  done

  # create checksums.txt
  pushd "${releaseDir}"
  for release in *; do
    echo "generate checksum: $release"
    sha256sum "$release" >> checksums.txt
  done
  popd

  rmdir ${workDir}
}

# Replace this with something that extracts git commit messages, e.g.
# see https://github.com/kubernetes-sigs/kustomize/blob/master/releasing/compile-changelog.sh
function createChangeLog {
  local changeLogFile=$1
cat <<EOF >$changeLogFile
Hi, I am placeholder content
for a TBD changelog.
EOF
}


function getReleaseTagOrDie {
  echo "hey0"
  local discard=$(git diff-index --quiet HEAD --)
  echo "hey1"
  if [[ $? != 0 ]]; then
    echo "The repo has open files; cannot release."
    exit 1
  fi
  echo "hey1a"
  local tagLatest=$(git describe --tags --abbrev=0)
  if [[ ! "$tagLatest" =~ "^v[0-9]" ]]; then
    echo "Invalid tag: $tagLatest"
    exit 1
  fi
  echo "hey2"
  local commitAtTag=$(git show-ref -s refs/tags/$tagLatest)
  echo "hey3"
  local commitLatest=$(git rev-parse --verify HEAD)
  echo "hey4"
  if [[ "$commitAtTag" != "$commitLatest" ]]; then
    echo "   tagLatest = $tagLatest"
    echo " commitAtTag = $commitAtTag"
    echo "commitLatest = $commitLatest"
    echo "Commit mismatch; release at $tagLatest aborted."
    exit 1
  fi
  echo "hey5"
  return $tagLatest
}


function createRelease {

  local gitTag=$(git describe --tags --abbrev=0)

  local changeLogFile=$(mktemp)
  local releaseArtifactDir=$(mktemp -d)

  createChangeLog $gitTag $changeLogFile

  buildBinaries $gitTag $releaseArtifactDir

  local artifacts=("$releaseArtifactDir"/*)

  gh release create $gitTag \
      --title $gitTag \
      --draft \
      --notes-file $changeLogFile \
      "${artifacts[@]}"
}

tag=$(getReleaseTagOrDie)

echo "tag = $tag"

# createRelease "$gitTag"
