.PHONY: build test clean docker

GO=PKG_CONFIG_PATH=/usr/local/zeromq-4.2.2/arm/lib/pkgconfig GOARCH=arm CC=arm-linux-gnueabihf-gcc CGO_ENABLED=1 CGO_CFLAGS="-g -O2 -I/usr/local/zeromq-4.2.2/arm/include" CGO_LDFLAGS="-g -O2 -L/usr/local/zeromq-4.2.2/arm/lib -L/usr/arm-linux-gnueabihf/lib -Wl,-rpath-link /usr/local/zeromq-4.2.2/arm/lib -Wl,-rpath-link /usr/arm-linux-gnueabihf/lib" go

GOAMD64=PKG_CONFIG_PATH=/usr/local/zeromq-4.2.2/amd64/lib/pkgconfig GOOS=linux GOARCH=amd64 go

MICROSERVICES=output/arm/export-homebridge

MICROSERVICESAMD64=output/amd64/export-homebridge

VERSION=$(shell cat ./VERSION)

BUILD_TIME   := $(shell date "+%F %T")

BUILD_NAME      := github.com/conthing

GOFLAGS=-ldflags "-X ${BUILD_NAME}/utils/common.Version=$(VERSION)\
 -X '${BUILD_NAME}/utils/common.BuildTime=${BUILD_TIME}'"

GIT_SHA=$(shell git rev-parse HEAD)

.PHONY: $(MICROSERVICES)

build_arm: $(MICROSERVICES)

output/arm/export-homebridge:
	$(GO) build $(GOFLAGS) -o $@ ./cmd/export-homebridge

test:
	$(GO) test ./... -cover

clean:
	rm -f $(MICROSERVICES)

.PHONY: $(MICROSERVICESAMD64)

build_amd64: $(MICROSERVICESAMD64)

output/amd64/export-homebridge:
	$(GOAMD64) build $(GOFLAGS) -o $@ ./cmd/export-homebridge

test_amd64:
	$(GOAMD64) test ./... -cover

clean_amd64:
	rm -f $(MICROSERVICESAMD64)
