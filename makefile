NAME := av-api
OWNER := byuoitav
PKG := github.com/${OWNER}/${NAME}
DOCKER_URL := docker.pkg.github.com
DOCKER_PKG := ${DOCKER_URL}/${OWNER}/${NAME}

# version:
# use the git tag, if this commit
# doesn't have a tag, use the git hash
COMMIT_HASH := $(shell git rev-parse --short HEAD)
TAG := $(shell git rev-parse --short HEAD)
ifneq ($(shell git describe --exact-match --tags HEAD 2> /dev/null),)
	TAG = $(shell git describe --exact-match --tags HEAD)
endif

PRD_TAG_REGEX := "v[0-9]+\.[0-9]+\.[0-9]+"
DEV_TAG_REGEX := "v[0-9]+\.[0-9]+\.[0-9]+-.+"

# go stuff
PKG_LIST := $(shell go list ${PKG}/...)

.PHONY: all deps build test test-cov lint clean docker deploy run

all: clean build

test:
	@go test -v ${PKG_LIST}

test-cov:
	@go test -coverprofile=coverage.txt -covermode=atomic ${PKG_LIST}

lint:
	@golangci-lint run --tests=false

deps:
	@echo Downloading dependencies...
	@go mod download

build: deps
	@mkdir -p dist

	@echo
	@echo Building ${NAME} for linux-amd64...
	@env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o dist/${NAME}-bin

	@echo
	@echo Building ${NAME} for linux-arm...
	@env CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -v -o dist/${NAME}-arm

	@cp version.txt dist/version.txt

	@echo
	@echo Build output is located in ./dist/.

run:
	@go run .

docker: clean build
ifeq (${COMMIT_HASH}, ${TAG})
	@echo Building dev containers with tag ${COMMIT_HASH}

	@echo Building container ${DOCKER_PKG}/${NAME}-dev:${COMMIT_HASH}
	@docker build -f dockerfile --build-arg NAME=${NAME} -t ${DOCKER_PKG}/${NAME}-dev:${COMMIT_HASH} dist

	@echo Building container ${DOCKER_PKG}/rpi-${NAME}-dev:${COMMIT_HASH}
	@docker build -f dockerfile-arm --build-arg NAME=${NAME} -t ${DOCKER_PKG}/rpi-${NAME}-dev:${COMMIT_HASH} dist
else ifneq ($(shell echo ${TAG} | grep -x -E ${DEV_TAG_REGEX}),)
	@echo Building dev containers with tag ${TAG}

	@echo Building container ${DOCKER_PKG}/${NAME}-dev:${TAG}
	@docker build -f dockerfile --build-arg NAME=${NAME} -t ${DOCKER_PKG}/${NAME}-dev:${TAG} dist

	@echo Building container ${DOCKER_PKG}/rpi-${NAME}-dev:${TAG}
	@docker build -f dockerfile-arm --build-arg NAME=${NAME} -t ${DOCKER_PKG}/rpi-${NAME}-dev:${TAG} dist
else ifneq ($(shell echo ${TAG} | grep -x -E ${PRD_TAG_REGEX}),)
	@echo Building prd containers with tag ${TAG}

	@echo Building container ${DOCKER_PKG}/${NAME}:${TAG}
	@docker build -f dockerfile --build-arg NAME=${NAME} -t ${DOCKER_PKG}/${NAME}:${TAG} dist

	@echo Building container ${DOCKER_PKG}/rpi-${NAME}:${TAG}
	@docker build -f dockerfile-arm --build-arg NAME=${NAME} -t ${DOCKER_PKG}/rpi-${NAME}:${TAG} dist
endif

deploy: docker
	@echo Logging into Github Package Registry
	@docker login ${DOCKER_URL} -u ${DOCKER_USERNAME} -p ${DOCKER_PASSWORD}

ifeq (${COMMIT_HASH}, ${TAG})
	@echo Pushing dev containers with tag ${COMMIT_HASH}

	@echo Pushing container ${DOCKER_PKG}/${NAME}-dev:${COMMIT_HASH}
	@docker push ${DOCKER_PKG}/${NAME}-dev:${COMMIT_HASH}

	@echo Pushing container ${DOCKER_PKG}/rpi-${NAME}-dev:${COMMIT_HASH}
	@docker push ${DOCKER_PKG}/rpi-${NAME}-dev:${COMMIT_HASH}
else ifneq ($(shell echo ${TAG} | grep -x -E ${DEV_TAG_REGEX}),)
	@echo Pushing dev containers with tag ${TAG}

	@echo Pushing container ${DOCKER_PKG}/${NAME}-dev:${TAG}
	@docker push ${DOCKER_PKG}/${NAME}-dev:${TAG}

	@echo Pushing container ${DOCKER_PKG}/rpi-${NAME}-dev:${TAG}
	@docker push ${DOCKER_PKG}/rpi-${NAME}-dev:${TAG}
else ifneq ($(shell echo ${TAG} | grep -x -E ${PRD_TAG_REGEX}),)
	@echo Pushing prd containers with tag ${TAG}

	@echo Pushing container ${DOCKER_PKG}/${NAME}:${TAG}
	@docker push ${DOCKER_PKG}/${NAME}:${TAG}

	@echo Pushing container ${DOCKER_PKG}/rpi-${NAME}:${TAG}
	@docker push ${DOCKER_PKG}/rpi-${NAME}:${TAG}
endif

clean:
	@go clean
	@rm -rf dist/
