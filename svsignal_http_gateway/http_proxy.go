package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"syscall"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	svsignal "github.com/lasaleks/svsignal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type HttpProxy struct {
	proxyAddr     string
	proxyUnixSock string
	grpcAddr      string
	listen        net.Listener
	server        *http.Server
}

func newUnixSocket(path string) (net.Listener, error) {
	if err := syscall.Unlink(path); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	/*mask := syscall.Umask(0777)
	defer syscall.Umask(0777)*/
	l, err := net.Listen("unix", path)
	if err != nil {
		return nil, err
	}
	if err := os.Chmod(path, 0777); err != nil {
		log.Printf("Error chmod unixsocket %s;%v", path, err)
	}
	return l, nil
}

func (h *HttpProxy) initUnixSocketServer() error {
	var err error
	if h.proxyAddr != "" {
		log.Printf("listen on %s", h.proxyAddr)
		h.listen, err = net.Listen("tcp", h.proxyAddr)
	} else {
		log.Printf("listen on %s", h.proxyUnixSock)
		h.listen, err = newUnixSocket(h.proxyUnixSock)
	}
	if err != nil {
		return err
	}
	return nil
}

// отправляемое сообщение клиенту
type WSM_MSG struct {
	WSM_TYPE string `json:"WSM_TYPE"`
	WSM_DATA string `json:"WSM_DATA"`
}

func (h *HttpProxy) run(wg *sync.WaitGroup) {
	defer wg.Done()
	grcpConn, err := grpc.Dial(
		h.grpcAddr,
		[]grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
		}...,
	)
	if err != nil {
		log.Fatalln("failed to connect to grpc", err)
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	grpcGWMux := runtime.NewServeMux()
	svsignal.RegisterSVSignalHandler(
		context.Background(),
		grpcGWMux,
		grcpConn,
	)

	h.server = &http.Server{Addr: h.proxyAddr, Handler: nil}
	h.server.Handler = grpcGWMux
	err = h.server.Serve(h.listen)
	if err == http.ErrServerClosed { // graceful shutdown
		log.Println("commencing Http server shutdown...")
		log.Println("Http server was shut down.")
	} else if err != nil {
		log.Printf("Http server error: %v\n", err)
	}
}

func (h *HttpProxy) close() {
	h.server.Close()
}
