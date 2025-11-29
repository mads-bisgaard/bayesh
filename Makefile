

.PHONY: clean
clean:
	git clean -fd

.PHONY: bats-tests
bats-tests:
	docker run \
		--user $(id -u):$(id -g) \
		-v "$(shell pwd):/code" \
		madsbis/bayesh-bats-testing:v5 \
		--print-output-on-failure \
		--verbose-run \
		tests
	
# VERSION is dynamically set from git. It will be vX.Y.Z for a tag,
# or vX.Y.Z-<commits>-g<hash> for a dev build.
VERSION := $(shell git describe --tags --always --dirty)
ARCH := $(shell go env GOARCH)
.PHONY: build
build:
	mkdir -p build
	cp -r bin/. build
	cp -r shell/. build
	go build -ldflags="-X 'main.version=${VERSION}'" -o ./build/bayesh ./main.go

.PHONY: release
release: build
	mkdir -p dist
	tar -czf dist/bayesh-$(VERSION)-linux-$(ARCH).tar.gz -C build .