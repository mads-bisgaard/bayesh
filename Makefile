

.PHONY: clean
clean:
	git clean -fd

.PHONY: bats-tests
bats-tests:
	docker run \
		-v "$(shell pwd):/code" \
		madsbis/bayesh-bats-testing:v5 \
		--print-output-on-failure \
		--verbose-run \
		tests
	
# VERSION is dynamically set from git. It will be vX.Y.Z for a tag,
# or vX.Y.Z-<commits>-g<hash> for a dev build.
ifdef BAYESH_VERSION
VERSION := $(BAYESH_VERSION)
else
VERSION := $(shell git describe --tags --always --dirty)
endif
ARCH := $(shell go env GOARCH)

.PHONY: build
build:
	mkdir -p build
	go build -ldflags="-X 'main.version=${VERSION}'" -o ./build/bayesh ./main.go

.PHONY: release
release: build
	grep -qF $(VERSION) install.sh
	mkdir -p build
	go build -ldflags="-X 'main.version=${VERSION}'" -o ./build/bayesh-$(VERSION)-linux-$(ARCH) ./main.go	