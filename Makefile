bin:
	go build $(BUILD_FLAGS) -o ./dn-cf-dns .

dev: BUILD_FLAGS = -tags "dev"
dev: bin

fmt:
	goimports -w .

test:
	go test ./...

vet:
	go vet ./...

.PHONY: bin dev fmt test vet
