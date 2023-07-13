
.PHONY: test build

build:
	go build ./...

test:
	go install ./cmd/... &&
	go generate . ./cmd/...  &&
	go test ./...
	go test -cover -race -timeout=2m -json ./... | tparse -skip
