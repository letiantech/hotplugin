#!/bin/sh
export GO111MODULE=on
go build -buildmode=plugin ./test/testplugin.go
go run ./test/test.go
