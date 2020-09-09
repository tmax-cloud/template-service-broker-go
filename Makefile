SDK	= operator-sdk

REGISTRY      ?= 172.22.11.2:30500
VERSION       ?= 0.0.1

PACKAGE_NAME  = github.com/jitaeyun/template-service-broker

SERVICE_BROKER_NAME  = tsb
SERVICE_BROKER_IMG   = $(REGISTRY)/$(SERVICE_BROKER_NAME):$(VERSION)
CLUSTER_SERVICE_BROKER_IMG = $(REGISTRY)/cluster-$(SERVICE_BROKER_NAME):$(VERSION)

BIN = ./build/_output/bin

.PHONY: build build-tsb build-cluster-tsb
build: build-tsb build-cluster-tsb

build-tsb:
	GOOS=linux CGO_ENABLED=0 go build -o $(BIN)/tsb/template-service-broker $(PACKAGE_NAME)/pkg/server/tsb
build-cluster-tsb:
	GOOS=linux CGO_ENABLED=0 go build -o $(BIN)/cluster-tsb/template-service-broker	 $(PACKAGE_NAME)/pkg/server/cluster-tsb

.PHONY: image image-tsb image-cluster-tsb
image: image-tsb image-cluster-tsb

image-tsb:
	docker build -f build/service-broker/tsb/Dockerfile -t $(SERVICE_BROKER_IMG) .
image-cluster-tsb:
	docker build -f build/service-broker/cluster-tsb/Dockerfile -t $(CLUSTER_SERVICE_BROKER_IMG) .

.PHONY: push push-tsb push-cluster-tsb
push: push-tsb push-cluster-tsb

push-tsb:
	docker push $(SERVICE_BROKER_IMG)
push-cluster-tsb:
	docker push $(CLUSTER_SERVICE_BROKER_IMG)
