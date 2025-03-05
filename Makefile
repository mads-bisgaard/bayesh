
.PHONY: bats-tests

bats-tests:
	docker run -v "$(shell pwd):/code" madsbis/bayesh-bats-testing:v2 --print-output-on-failure tests
