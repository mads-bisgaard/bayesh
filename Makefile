

.PHONY: install
install:
	uv pip install -e .[dev]

.PHONY: bats-tests
bats-tests:
	docker run -v "$(shell pwd):/code" madsbis/bayesh-bats-testing:v2 --print-output-on-failure tests

.PHONY: build
build:
	mkdir -p build
	go build -o ./build/bayesh ./main.go