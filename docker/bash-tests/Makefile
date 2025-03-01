
repo 		:= madsbis
imagename	:= bayesh-bats-testing
version 	:= v2

.PHONY: build push

build:
	docker build -t $(repo)/$(imagename):$(version) .

push:
	docker push $(repo)/$(imagename):$(version)