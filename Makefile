APP_VERSION = $(shell git describe --abbrev=0 --tags)
GIT_COMMIT = $(shell git rev-parse --short HEAD)
BUILD_DATE = $(shell date -u "+%Y%m%d-%H%M")
VERSION_PKG = github.com/InjectiveLabs/injective-liquidator-bot/version
IMAGE_NAME := public.ecr.aws/l9h3g6c6/injective-liquidator-bot

all:

image:
	docker build --build-arg GIT_COMMIT=$(GIT_COMMIT) -t $(IMAGE_NAME):local -f Dockerfile .
	docker tag $(IMAGE_NAME):local $(IMAGE_NAME):$(GIT_COMMIT)
	docker tag $(IMAGE_NAME):local $(IMAGE_NAME):latest

push:
	docker push $(IMAGE_NAME):$(GIT_COMMIT)
	docker push $(IMAGE_NAME):latest

install: export GOPROXY=direct
install: export VERSION_FLAGS="-X $(VERSION_PKG).GitCommit=$(GIT_COMMIT) -X $(VERSION_PKG).BuildDate=$(BUILD_DATE)"
install:
	go install \
		-ldflags $(VERSION_FLAGS) \
		./cmd/...

.PHONY: install build image push test gen

build:
	go build -o injective-labs-liquidator ./cmd/injective-liquidator-bot/main.go ./cmd/injective-liquidator-bot/liquidator.go ./cmd/injective-liquidator-bot/metrics.go ./cmd/injective-liquidator-bot/options.go ./cmd/injective-liquidator-bot/util.go

test:
	# go clean -testcache
	go test ./test/...
