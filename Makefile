TAG = $(shell git describe --tags --abbrev=0)
IMAGE_VERSION = $(shell echo ${TAG} | cut -c2-)

.PHONY: build
build-prod:
	sh hack/build.sh prod

.PHONY: docker
docker:
	docker build -t rings-node:${IMAGE_VERSION} -f ./docker/Dockerfile .

clean:
	@rm -rf ./builds
