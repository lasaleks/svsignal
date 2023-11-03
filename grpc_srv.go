package main

import (
	"fmt"
	"log"
	"net"
	"sync"

	svsignal "github.com/lasaleks/svsignal/proto"
	"google.golang.org/grpc"
)

// grpc server
type gRPCServer struct {
	Addr   string
	server *grpc.Server
}

func (s *gRPCServer) run(wg *sync.WaitGroup) {
	defer wg.Done()
	lis, err := net.Listen("tcp", s.Addr)
	if err != nil {
		log.Fatalln("cant listet port", err)
	}

	s.server = grpc.NewServer()
	svsignal.RegisterSVSignalServer(s.server, savesignal)

	fmt.Printf("starting grpc server at %s\n", s.Addr)
	s.server.Serve(lis)
	fmt.Println("end gRPCServer.run")
}

func (s *gRPCServer) Close() {
	s.server.Stop()
}
