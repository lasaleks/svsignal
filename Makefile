MUTABLE_VERSION := latest
VERSION := $(shell git describe --tags)

BUILD_DATE := $(shell date -u +%d%m%y.%H%M%S)
DOCKER_IMAGE := svsignal
IMAGE_STAGING := dregistryserver:5043/svsignal
 
docker: 
	docker build --build-arg "version=$(VERSION)" --build-arg "build_date=$(BUILD_DATE)" --tag=$(DOCKER_IMAGE):$(VERSION) --file Dockerfile .

docker-staging: docker
	docker tag $(DOCKER_IMAGE):$(VERSION) $(IMAGE_STAGING):$(VERSION)
	docker tag $(DOCKER_IMAGE):$(VERSION) $(IMAGE_STAGING):$(MUTABLE_VERSION)

push-staging: docker-staging
	docker push $(IMAGE_STAGING):$(VERSION)
	docker push $(IMAGE_STAGING):$(MUTABLE_VERSION)

run:
	go run . --pid-file /tmp/svsignal.pid --config-file etc/config.yaml
