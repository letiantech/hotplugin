#!/bin/sh
export GO111MODULE=on
go build -buildmode=plugin ./testplugin.go
go run ./main.go
