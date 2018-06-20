IMAGE_TAG=v1alpha1
QUAY_PASS?=biggestsecret
CRAWLER_URL?=https://docs.docker.com
CRAWLER_DEPTH?=3

compile:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o creepycrawly .

dev: compile
	docker build -t quay.io/tamarakaufler/creepycrawly:$(IMAGE_TAG) .

build: dev
	docker login quay.io -u tamarakaufler -p $(QUAY_PASS)
	docker push quay.io/tamarakaufler/creepycrawly:$(IMAGE_TAG)

run:
	docker run \
	--name=creepycrawly \
	--rm \
	quay.io/tamarakaufler/creepycrawly:$(IMAGE_TAG) \
	-url=$(CRAWLER_URL) \
	-depth=$(CRAWLER_DEPTH)
