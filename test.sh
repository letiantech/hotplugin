#!/bin/sh
go build -buildmode=plugin ./test/testplugin.go
go run ./test/test.go
