
.PHONY: build
build:
	go build -o compound.local

.PHONY: build-dev
build-dev:
	sh scripts/build.sh dev

.PHONY: build-prod
build-prod:
	sh scripts/build.sh prod


clean:
	rm -rf compound.*
