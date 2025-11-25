

.PHONY: install
install:
	uv pip install -e .[dev]

.PHONY: bats-tests
bats-tests:
	docker run -v "$(shell pwd):/code" madsbis/bayesh-bats-testing:v2 --print-output-on-failure tests
	
# VERSION is dynamically set from git. It will be vX.Y.Z for a tag,
# or vX.Y.Z-<commits>-g<hash> for a dev build.
VERSION := $(shell git describe --tags --always --dirty)
.PHONY: build
build:
	mkdir -p build/bin build/shell
	cp -r bin/. build/bin/
	cp -r shell/. build/shell/
	go build -ldflags="-X 'main.version=${VERSION}'" -o ./build/bin/bayesh ./main.go
	