
.PHONY: bats-tests

bats-tests:
	docker run -it -v "$(shell pwd):/code" madsbis/bayesh-bats-testing:v0 tests
