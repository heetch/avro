#!/bin/sh

set -ex

go install ./cmd/avro-generate-go
go generate ./...
go test ./...
