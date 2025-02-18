# Installation options

[release from GitHub]: https://github.com/monopole/mdrip/releases
[Go tool]: https://golang.org/dl
[wsl]: https://learn.microsoft.com/en-us/windows/wsl

## Install via the [Go tool]

```
go install github.com/monopole/mdrip/v2@latest
```

## Download a [release from GitHub]

In a linux or darwin bash shell:
```
tag=v2.0.2          # or some other release tag
os=linux            # or darwin
arch=amd64          # or arm64
```
Download and unpack
```
file=mdrip_${tag}_${os}_${arch}.tar.gz
wget -q https://github.com/monopole/mdrip/releases/download/${tag}/${file}
tar -xf $file
rm $file
./mdrip version     # confirm the release tag
```
Put it on your `PATH`, e.g.
```
mv ./mdrip ~/go/bin
```

On Windows, basic code block extraction works via the `print` command.
But the `test` command won't work, as it's currently hardcoded to 
run code blocks under `bash`, not `powershell`. The `test` command could be
made to work with `powershell` with some new command line options and/or
OS detection. However, the `tmux` integration with the `serve` command won't
work since `tmux` isn't supported on Windows. Nevertheless, a Windows binary is
released and can be installed via:
```
$tag="v2.0.0-rc12"  # or some other release tag
$arch="amd64"       # or arm64

$file="mdrip_${tag}_windows_${arch}.zip"
$url="https://github.com/monopole/mdrip/releases/download/${tag}/${file}"
Invoke-WebRequest -Uri ${url} -Outfile ${file}
Expand-Archive ${file} -Force -DestinationPath .
rm ${file}

./mdrip version     # Visually confirm the release tag.
```
FWIW, `mdrip` works fine under [wsl], but obviously that's linux, not Windows.

[dockerhub]: https://hub.docker.com/repository/docker/monopole/mdrip/tags

## Run a container image from [dockerhub]

```
image=monopole/mdrip:latest
```
```
docker run $image version
```
```
docker run $image help
```

Do basic extraction of all the code blocks below your current directory:
```
dir=$(pwd)
docker run \
  --mount type=bind,source=$dir,target=/mnt \
  $image print /mnt
```

The `test` command should also work, running over the markdown
in the mounted volume, failing on an error.

Using the `serve` command to serve rendered markdown from your mounted 
volume is possible this way, but the `tmux` integration - which is the
only thing interesting about serving markdown from `mdrip` - won't work
because the container does not have `tmux` installed.
The use case of running `tmux` and `mdrip serve` _in a container_,
presumably to manipulate necessarily ephemeral container state,
should be possible - but it's an odd use case to support.
