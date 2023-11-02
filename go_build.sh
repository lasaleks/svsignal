#!/bin/bash

VERSION=`git describe --tags`
date=`date -u +%d%m%y.%H%M%S`
go build -a -ldflags="-X main.VERSION=$VERSION -X main.BUILD=$date" -o ../bin/svsignal.bin
