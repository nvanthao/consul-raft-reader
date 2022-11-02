VERSION="0.1.0"
DOCKER_IMAGE="nvanthao/consul-raft-reader"

.PHONY: tidy
tidy: 
	@echo "--> Tidy modules"
	@go mod tidy

.PHONY: docker-build-local
docker-build-local:
	@echo "--> Build docker image"
	@docker build -t $(DOCKER_IMAGE) .

.PHONY: docker-publish
docker-publish: docker-build-local
	@echo "---> Tag docker image"
	@docker tag $(DOCKER_IMAGE) $(DOCKER_IMAGE):$(VERSION)
	@docker tag $(DOCKER_IMAGE) $(DOCKER_IMAGE):latest
	@echo "--> Publish docker image"
	@docker push -a $(DOCKER_IMAGE)

.PHONY: install
install: tidy
	@echo "--> Install binary"
	@go install .