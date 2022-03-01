# These variables get inserted into ./build/commit.go
BUILD_TIME=$(shell date)
GIT_REVISION=$(shell git rev-parse --short HEAD)
GIT_DIRTY=$(shell git diff-index --quiet HEAD -- || echo "modified-")

ldflags= -X gitlab.com/scpcorp/webwallet/build.GitRevision=${GIT_DIRTY}${GIT_REVISION} \
-X "gitlab.com/scpcorp/webwallet/build.BuildTime=${BUILD_TIME}"

racevars= history_size=3 halt_on_error=1 atexit_sleep_ms=2000

# all will build and install release binaries
all: release

# count says how many times to run the tests.
count = 1

# pkgs changes which packages the makefile calls operate on. run changes which
# tests are run during testing.
pkgs = \
	./cmd/scp-webwallet \
	./resources \
	./server

# release-pkgs determine which packages are built for release and distrubtion
# when running a 'make release' command.
release-pkgs = ./cmd/scp-webwallet

# run determines which tests run when running any variation of 'make test'.
run = .

# dependencies list all packages needed to run make commands used to build, test
# and lint the binaries locally and in CI systems.
dependencies:
	go get -d ./...
	./install-dependencies.sh

# fmt calls go fmt on all packages.
fmt:
	gofmt -s -l -w $(pkgs)

# vet calls go vet on all packages.
# NOTE: go vet requires packages to be built in order to obtain type info.
vet:
	go vet $(pkgs)

# lint runs golangci-lint.
lint:
	golangci-lint run -c .golangci.yml ./...

# debug builds and installs debug binaries. This will also install the utils.
debug:
	go install -tags='debug profile netgo' -ldflags='$(ldflags)' $(pkgs)
debug-race:
	GORACE='$(racevars)' go install -race -tags='debug profile netgo' -ldflags='$(ldflags)' $(pkgs)

# dev builds and installs developer binaries. This will also install the utils.
dev:
	go install -tags='dev debug profile netgo' -ldflags='$(ldflags)' $(pkgs)
dev-race:
	GORACE='$(racevars)' go install -race -tags='dev debug profile netgo' -ldflags='$(ldflags)' $(pkgs)

# release builds and installs release binaries.
release:
	go install -tags='netgo' -ldflags='-s -w $(ldflags)' $(release-pkgs)
release-race:
	GORACE='$(racevars)' go install -race -tags='netgo' -ldflags='-s -w $(ldflags)' $(release-pkgs)
release-util:
	go install -tags='netgo' -ldflags='-s -w $(ldflags)' $(release-pkgs) $(util-pkgs)

# clean removes all directories that get automatically created during
# development.
clean:
ifneq ("$(OS)","Windows_NT")
# Linux
	rm -rf cover
else
# Windows
	- DEL /F /Q cover
endif

test:
	go test -short -tags='debug testing netgo' -timeout=60s $(pkgs) -run=$(run) -count=$(count)
cover: clean
	@mkdir -p cover
	@for package in $(pkgs); do                                                                                                                                 \
		mkdir -p `dirname cover/$$package`                                                                                                                      \
		&& go test -tags='testing debug netgo' -timeout=500s -covermode=atomic -coverprofile=cover/$$package.out ./$$package -run=$(run) || true \
		&& go tool cover -html=cover/$$package.out -o=cover/$$package.html ;                                                                                    \
	done
.PHONY: all fmt vet lint release clean test cover
