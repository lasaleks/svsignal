#!/bin/sh
#go build -o bin/diagemn.bin main.go base.go diag.go data.go config.go grpc_srv.go service.go http_proxy.go cmd_test_signal.go hub.go

VERSION=`git describe --tags`
date=`date -u +%d%m%y.%H%M%S`
CGO_ENABLED=0 go build -a -ldflags="-X main.VERSION=$VERSION -X main.BUILD=$date" -o ../../bin/svsignal_http_gateway.bin
