#!/usr/bin/env bash
set -e

# version and keys are supplied as arguments
version="$1"
rc=`echo $version | awk -F - '{print $2}'`
if [[ -z $version ]]; then
	echo "Usage: $0 VERSION"
	exit 1
fi

# setup build-time vars
sharedldflags="-s -w -X 'gitlab.com/scpcorp/webwallet/build.GitRevision=`git rev-parse --short HEAD`' -X 'gitlab.com/scpcorp/webwallet/build.BuildTime=`git show -s --format=%ci HEAD`' -X 'gitlab.com/scpcorp/webwallet/build.ReleaseTag=${rc}'"

function build {
  os=$1
  arch=$2
  echo Building ${os}-${arch}...
  # create workspace
  product=scp-webwallet
  folder=release/$product-$version-$os-$arch
  rm -rf $folder
  mkdir -p $folder
  # compile and hash binaries
  pkg=scp-webwallet
  bin=${pkg}
  if [ "$os" == "windows" ]; then
    bin=${bin}.exe
  fi
  ldflags=$sharedldflags
  # Appify the scp-webwallet windows release. More documentation at `./release-scripts/app_resources/windows/README.md`
  if [ "$os" == "windows" ]; then
    # copy metadata files to the package
    cp ./release-scripts/app_resources/windows/rsrc_windows_386.syso ./cmd/scp-webwallet/rsrc_windows_386.syso
    cp ./release-scripts/app_resources/windows/rsrc_windows_amd64.syso ./cmd/scp-webwallet/rsrc_windows_amd64.syso
    # on windows build an application binary instead of a command line binary.
    ldflags="$sharedldflags -H windowsgui"
  fi
  # Appify the scp-webwallet darwin release. More documentation at `./release-scripts/app_resources/darwin/README.md`.
  if [ "$os" == "darwin" ]; then
    # copy the scp-webwallet.app template container into the release directory
    cp -a ./release-scripts/app_resources/darwin/scp-webwallet.app $folder/$bin.app
    # touch the scp-webwallet.app container to reset the time created timestamp
    touch $folder/$bin.app
    # set the build target to be inside the scp-webwallet.app container
    bin=scp-webwallet.app/Contents/MacOS/scp-webwallet.app
  fi
  GOOS=${os} GOARCH=${arch} go build -a -tags 'netgo' -trimpath -ldflags="$ldflags" -o $folder/$bin ./cmd/$pkg
  # Cleanup scp-webwallet windows release.
  if [ "$os" == "windows" ]; then
    rm ./cmd/scp-webwallet/rsrc_windows_386.syso
    rm ./cmd/scp-webwallet/rsrc_windows_amd64.syso
  fi
  (
    cd release/
    sha1sum $product-$version-$os-$arch/$bin >> $product-$version-SHA1SUMS.txt
  )
  cp -r LICENSE README.md $folder
}

# Build amd64 binaries.
for os in darwin linux windows; do
  build "$os" "amd64"
done

# Build Raspberry Pi binaries.
build "linux" "arm64"
