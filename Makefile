
# .PHONY: build
# build:
# 	go build -o price-oracle.local

# .PHONY: build-dev
# build-dev:
# 	sh scripts/build.sh dev

# .PHONY: build-prod
# build-prod:
# 	sh scripts/build.sh prod


# clean:
# 	rm -rf price-oracle.*

COMMIT = $(shell git rev-parse --short HEAD)
VERSION = $(shell git describe --abbrev=0) 

ENV = $*
GO = go

%-local: GO = GO111MODULE=on go
%-test: GO = GOOS=linux GOARCH=amd64 GO111MODULE=on go
%-prod: GO = GOOS=linux GOARCH=amd64 GO111MODULE=on go

.PHONY: clean
clean-%:	
	@echo "cleaning building caches and configs......................."
	${GO} clean
	rm -f ./compound
	rm -f ./compound.${ENV}
	rm -f ./config/config.yaml

sync-%: 
	@echo "sync code and config file..........................."
	git pull
	cp -f ./deploy/config.${ENV}.yaml ./config/config.yaml

build-%: clean-% sync-%
	${GO} build --ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -o compound.${ENV}"
	cp compound.${ENV} compound

docker-build-%: build-%
	docker build -t 945908130943.dkr.ecr.ap-northeast-1.amazonaws.com/compound-${ENV}:${VERSION} . 

.PHONY: aws-login
aws-login:
	$(shell aws ecr get-login --no-include-email --region ap-northeast-1)

deploy-%: docker-build-%
	docker push 945908130943.dkr.ecr.ap-northeast-1.amazonaws.com/compound-${ENV}:${VERSION}
