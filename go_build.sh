#!/bin/bash

VERSION=`git describe --tags`
date=`date -u +%d%m%y.%H%M%S`
CGO_ENABLED=0 go build -a -ldflags="-X main.VERSION=$VERSION -X main.BUILD=$date" -o bin/svsignal
