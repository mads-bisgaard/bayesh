

.PHONY: clean
clean:
	git clean -fd

.PHONY: bats-tests
bats-tests: build
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

MANUAL_TEST_DIR := $(CURDIR)/build/manual-test
MANUAL_TEST_TMUX := tmux -L bayesh-manual-test -f /dev/null

# start a detached tmux session running the local build with a fresh database
.PHONY: manual-test
manual-test: manual-test-stop build
	rm -rf $(MANUAL_TEST_DIR)
	mkdir -p $(MANUAL_TEST_DIR)
	cp tests/manual-test.zshrc $(MANUAL_TEST_DIR)/.zshrc
	PATH="$(CURDIR)/build:$$PATH" \
	ZDOTDIR=$(MANUAL_TEST_DIR) \
	BAYESH_DIR=$(MANUAL_TEST_DIR) \
	$(MANUAL_TEST_TMUX) new-session -d -s bayesh -x 120 -y 30 zsh -i

.PHONY: manual-test-stop
manual-test-stop:
	$(MANUAL_TEST_TMUX) kill-server 2> /dev/null || true

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
