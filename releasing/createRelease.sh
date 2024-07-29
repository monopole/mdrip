#!/bin/bash

# Always declare vars.
set -u

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
        tar cvfz ${artifact}.tar.gz --directory ${workDir} $binaryName
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

Release notes placeholder.

EOF
}


function getReleaseTagOrDie {
  local discard=$(git diff-index --quiet HEAD --)
  if [ $? -ne 0 ]; then
    echo "ERR: The repo has open files; cannot release."
    return
  fi
  local tagLatest=$(git describe --tags --abbrev=0)
  echo "wut $?"
  if [ $? -ne 0 ]; then
    echo "ERR: The are no tags in this repo.  Apply a tag to release."
    return
  fi
  if [[ ! "$tagLatest" =~ ^v[0-9]\. ]]; then
    echo "ERR: Invalid tag pattern: $tagLatest"
    return
  fi
  local commitAtTag=$(git show-ref -s refs/tags/$tagLatest)
  local commitLatest=$(git rev-parse --verify HEAD)
  if [[ "$commitAtTag" != "$commitLatest" ]]; then
    echo "ERR: This repo changed after application of tag: $tagLatest"
    echo "               commit at that tag: $commitAtTag"
    echo "     does not match latest commit: $commitLatest"
    echo "You probably want to apply a new tag."
    return
  fi
  echo "$tagLatest"
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


# set -x
# set -o errexit
# set -o nounset
# set -o pipefail
# set +e
tag=$(getReleaseTagOrDie)

if [[ "$tag" =~ ^ERR: ]]; then
  echo "$tag"
  exit 1
fi

echo "The tag is $tag"
# createRelease "$gitTag"
