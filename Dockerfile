FROM golang:1.21.3

WORKDIR /build/

COPY go.mod go.sum ./

RUN go mod download


FROM golang:1.21.3 AS builder

WORKDIR /build

ADD go.mod .

COPY . .

#RUN go build -o hello hello.go

RUN env GOOS=linux GOARCH=amd64 go build -a -ldflags="-X main.VERSION=$VERSION -X main.BUILD=$date" -o svsignal.bin

FROM alpine

WORKDIR /build

COPY --from=builder /build/svsignal.bin /build/svsignal.bin

CMD [". /svsignal.bin --version"]
