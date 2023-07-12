
.PHONY: test build

build:
	go build ./...

test:
	go test -cover -race -timeout=2m -json ./... | tparse -skip

