#!/usr/bin/env bash
set -e

# version and keys are supplied as arguments
version=${1}
keyfile=${2} # optional
rc=${version}
if [[ -z ${version} ]]; then
	echo "Usage: ${0} VERSION OPTIONAL_KEYFILE"
	exit 1
fi
if [[ -z ${keyfile} ]]; then
  echo "No gpg keyfile was supplied. Binaries will not be signed."
else
  rc=`echo ${version} | awk -F - '{print $2}'`
  gpg --import ${keyfile}
fi

# setup build-time vars
sharedldflags="-s -w -X 'gitlab.com/scpcorp/webwallet/build.GitRevision=`git rev-parse --short HEAD`' -X 'gitlab.com/scpcorp/webwallet/build.BuildTime=`git show -s --format=%ci HEAD`' -X 'gitlab.com/scpcorp/webwallet/build.ReleaseTag=${rc}'"

function build {
  os=${1}
  arch=${2}
  pkg=${3}
  releasePkg=${pkg}-${version}-${os}-${arch}
  echo Building ${releasePkg}...
  # set binary name
  bin=${pkg}
  if [ ${os} == "windows" ]; then
    bin=${bin}.exe
  elif [ ${os} == "darwin" ]; then
    bin=${bin}.app
  fi
  # set folder
  folder=release/${releasePkg}
  rm -rf ${folder}
  mkdir -p ${folder}
  # set path to bin
  binpath=${folder}/${bin}
  if [ ${os} == "darwin" ]; then
    # Appify the scp-webwallet darwin release. More documentation at `./release-scripts/app_resources/darwin/scp-webwallet/README.md`.
    cp -a ./release-scripts/app_resources/darwin/${pkg}/${bin} ${binpath}
    # touch the scp-webwallet.app container to reset the time created timestamp
    touch ${binpath}
    binpath=${binpath}/Contents/MacOS/${bin}
  fi
  # set ldflags
  ldflags=$sharedldflags
  if [ ${os} == "windows" ]; then
    # Appify the scp-webwallet windows release. More documentation at `./release-scripts/app_resources/windows/scp-webwallet/README.md`
    cp ./release-scripts/app_resources/windows/${pkg}/rsrc_windows_386.syso ./cmd/${pkg}/rsrc_windows_386.syso
    cp ./release-scripts/app_resources/windows/${pkg}/rsrc_windows_amd64.syso ./cmd/${pkg}/rsrc_windows_amd64.syso
    # on windows build an application binary instead of a command line binary.
    ldflags="${sharedldflags} -H windowsgui"
  fi
  GOOS=${os} GOARCH=${arch} go build -a -tags 'netgo' -trimpath -ldflags="${ldflags}" -o ${binpath} ./cmd/${pkg}
  # Cleanup scp-webwallet windows release.
  if [ ${os} == "windows" ]; then
    rm ./cmd/${pkg}/rsrc_windows_386.syso
    rm ./cmd/${pkg}/rsrc_windows_amd64.syso
  fi
  if ! [[ -z ${keyfile} ]]; then
    gpg --armour --output ${folder}/${bin}.asc --detach-sig ${binpath}
  fi
  sha1sum ${binpath} >> release/${pkg}-${version}-SHA1SUMS.txt
  cp -r LICENSE README.md ${folder}  
  (
    cd release/
    zip -rq ${releasePkg}.zip ${releasePkg}
  )
  if ! [[ -z ${keyfile} ]]; then
    gpg --armour --output ${folder}.zip.asc --detach-sig ${folder}.zip
  fi
}

# Build amd64 binaries.
for os in darwin linux windows; do
  build ${os} "amd64" "scp-webwallet"
done

# Build Raspberry Pi binaries.
build "linux" "arm64" "scp-webwallet"
