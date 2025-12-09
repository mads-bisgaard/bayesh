

.PHONY: clean
clean:
	git clean -fd

.PHONY: bats-tests
bats-tests:
	docker run \
		-v "$(shell pwd):/code" \
		-v "$(shell pwd)/build:/usr/local/bin" \
		madsbis/bayesh-bats-testing:v6 \
		--print-output-on-failure \
		--verbose-run \
		tests
	
ifdef BAYESH_VERSION
VERSION := $(BAYESH_VERSION)
else
VERSION := $(shell git describe --tags --always --dirty)
endif

.PHONY: build
build:
	rm -rf build && mkdir -p build
	go build -ldflags="-X 'main.version=${VERSION}'" -o ./build/bayesh ./main.go

# supported architectures
ARCH := amd64 arm

.PHONY: release
release:
	grep -qF $(VERSION) install.sh
	rm -rf dist && mkdir -p dist
	for arc in $(ARCH); do \
		tmpdir=$$(mktemp -d); \
		GOARCH=$$arc go build -ldflags="-s -w -X 'main.version=${VERSION}'" -o $$tmpdir/bayesh ./main.go; \
		tar -czf ./dist/bayesh-$(VERSION)-linux-$$arc.tar.gz -C $$tmpdir bayesh; \
		rm -rf $$tmpdir; \
	done
