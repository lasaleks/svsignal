FROM golang:1.21.3 AS builder

WORKDIR /build
COPY . .
# env GOOS=linux GOARCH=amd64 
ARG version=unknown
ARG build_date=unknown
RUN go build -a -ldflags="-X main.VERSION=$version -X main.BUILD=$build_date" -o svsignal.bin

FROM frolvlad/alpine-glibc

COPY --from=builder /build/svsignal.bin /app/svsignal/bin/svsignal.bin

COPY etc/svsignal_prod.yaml /app/svsignal/etc/svsignal.yaml
COPY etc/server.yaml /app/svsignal/etc/server.yaml

COPY py/ /app/svsignal/py
COPY templates /app/svsignal/templates
COPY static /app/svsignal/static

VOLUME /app/svsignal/database/
VOLUME /sock/
WORKDIR /app/svsignal
CMD ./bin/svsignal.bin --config-file /app/svsignal/etc/svsignal.yaml --pid-file /tmp/service.pid
