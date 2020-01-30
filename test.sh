#!/bin/sh

set -ex
PKG=github.com/heetch/avro
go install $PKG/cmd/avro-generate-go
go generate $PKG/...
go test $PKG/...
