DOCKER ?= podman

download:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.42.0
lint:
	./bin/golangci-lint run --fix

default: clean build

build:
	go build -o ping-test ./cmd/web/main.go
clean:
	rm -rf ping-test
install:
	sudo cp ./ping-test /usr/local/bin/ping-test && chmod 755 /usr/local/bin/ping-test
docker-build:
	$(DOCKER)  build -t ghcr.io/rtdev7690/ping-test:latest . && $(DOCKER) push ghcr.io/rtdev7690/ping-test:latest
