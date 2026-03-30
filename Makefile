VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
BUILD_FLAGS += -trimpath -ldflags "-X main.version=$(VERSION)"

ALL_LINUX = linux-amd64 \
			linux-arm64

bin:
	go build $(BUILD_FLAGS) -o ./dn-cf-dns .

dist/%/dn-cf-dns:
	GOOS=$(firstword $(subst -, , $*)) \
		GOARCH=$(word 2, $(subst -, ,$*)) $(GOENV) \
		go build $(BUILD_FLAGS) -o $@ .

release: $(ALL_LINUX:%=dist/%/dn-cf-dns)

clean:
	rm -r dist

dev: BUILD_FLAGS = -tags "dev"
dev: bin

fmt:
	goimports -w .

test:
	go test $(TEST_FLAGS) $(shell go list ./...)

testv: TEST_FLAGS += -v
testv: test

vet:
	go vet ./...

.PHONY: bin clean dev fmt test testv vet
