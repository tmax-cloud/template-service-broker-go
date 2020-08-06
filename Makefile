SDK	= operator-sdk

REGISTRY      ?= 172.22.11.2:30500
VERSION       ?= 0.0.1

PACKAGE_NAME  = github.com/jwkim1993/template-service-broker

SERVICE_BROKER_NAME  = tsb
SERVICE_BROKER_IMG   = $(REGISTRY)/$(SERVICE_BROKER_NAME):$(VERSION)

BIN = ./build/_output/bin

.PHONY: build build-tsb
build: build-tsb

build-tsb:
	GOOS=linux CGO_ENABLED=0 go build -o $(BIN)/template-service-broker $(PACKAGE_NAME)/pkg/server

.PHONY: image image-tsb
image: image-tsb

image-tsb:
	docker build -f build/service-broker/Dockerfile -t $(SERVICE_BROKER_IMG) .

.PHONY: push push-tsb
push: push-tsb

push-tsb:
	docker push $(SERVICE_BROKER_IMG)
