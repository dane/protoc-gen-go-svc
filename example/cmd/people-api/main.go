package main

import (
	"flag"
	"log"
	"net"

	"google.golang.org/grpc"

	privatepb "github.com/dane/protoc-gen-go-svc/example/proto/go/private"
	servicepb "github.com/dane/protoc-gen-go-svc/example/proto/go/service"
	private "github.com/dane/protoc-gen-go-svc/example/service/private"
)

func main() {
	addr := flag.String("addr", ":8000", "service address")
	flag.Parse()

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}

	impl := &private.Service{Store: make(map[string]*privatepb.Person)}
	srv := grpc.NewServer()
	servicepb.RegisterServer(srv, impl)

	log.Printf("listening on address: %s", ln.Addr())
	if err := srv.Serve(ln); err != nil {
		log.Fatal(err)
	}
}
