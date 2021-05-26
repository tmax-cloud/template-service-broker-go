REGISTRY      ?= tmaxcloudck
VERSION       ?= 0.0.9

SERVICE_BROKER_IMG   = $(REGISTRY)/tsb:$(VERSION)
CLUSTER_SERVICE_BROKER_IMG = $(REGISTRY)/cluster-tsb:$(VERSION)

.PHONY: build push

# Build the docker image
build:
	docker build -f build/tsb/Dockerfile -t $(SERVICE_BROKER_IMG) .
	docker build -f build/cluster-tsb/Dockerfile -t $(CLUSTER_SERVICE_BROKER_IMG) . 

# Push the docker image 
push:
	docker push $(SERVICE_BROKER_IMG)
	docker push $(CLUSTER_SERVICE_BROKER_IMG)
