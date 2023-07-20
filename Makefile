
.PHONY: test build

build:
	go build ./...

test:
	go generate . ./cmd/...
	go test ./... -cover -race -timeout=2m -json ./... | tparse
