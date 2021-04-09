COMMIT = $(shell git rev-parse --short HEAD)
VERSION = $(shell git describe --abbrev=0) 

REPOSITORY_PATH = $(shell cat .config.ini)
ENV = $*
GO = GO111MODULE=on CGO_ENABLED=1 CGO_CFLAGS='-O -D__BLST_PORTABLE__' go

clean-%:	
	@echo "cleaning building caches and configs......................."
	${GO} clean
	rm -f ./compound
	rm -f ./compound.${ENV}
	rm -f ./config/config.yaml

sync-%: 
	@echo "sync code and config file..........................."
	# git pull
	cp -f ./deploy/config.${ENV}.yaml ./config/config.yaml

refresh-%: clean-% sync-%
	@echo "clean and sync ..............."

build: 
	${GO} build --ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT}"

docker-build-%: clean-% sync-%
	@echo "repository path -> ${REPOSITORY_PATH}"
	docker build -t ${REPOSITORY_PATH}/compound-${ENV}:${VERSION} -f ./deploy/docker/Dockerfile . 

.PHONY: aws-login
aws-login:
	$(shell aws ecr get-login --no-include-email --region ap-northeast-1)

deploy-%: docker-build-%
	docker push ${REPOSITORY_PATH}/compound-${ENV}:${VERSION}