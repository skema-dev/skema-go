package main

import (
	"context"
	"log"

	pb "github.com/skema-dev/skema-go/test/api/skema/test"
)

type rpcTestServer struct {
	pb.UnimplementedTestServer
}

// NewServer: Create new grpc server instance
func NewServer() pb.TestServer {
	svr := &rpcTestServer{
		// init custom fileds
	}
	return svr
}

// Heathcheck
func (s *rpcTestServer) Heathcheck(
	ctx context.Context,
	req *pb.HealthcheckRequest,
) (rsp *pb.HealthcheckResponse, err error) {
	// implement business logic here ...
	// ...

	log.Printf("Received from Heathcheck request: %v", req)
	rsp = &pb.HealthcheckResponse{
		Result: "health check ok",
	}

	return rsp, err
}

// Helloworld
func (s *rpcTestServer) Helloworld(ctx context.Context, req *pb.HelloRequest) (rsp *pb.HelloReply, err error) {
	// implement business logic here ...
	// ...

	log.Printf("Received from Helloworld request: %v", req)
	rsp = &pb.HelloReply{
		Msg:  "Hello world",
		Code: "0",
	}
	return rsp, err
}
