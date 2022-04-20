package server

import (
	"context"
	"fmt"
	"log"

	"grpc-dao/internal/dao"

	"github.com/skema-dev/skema-go/data"
	pb "github.com/skema-dev/skema-go/sample/api/skema/test"

	"github.com/google/uuid"
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

	result := ""

	user := data.Manager().GetDAO(&dao.User{})
	err = user.Upsert(&dao.User{
		UUID: uuid.New().String(),
		Name: req.Msg,
	}, nil, nil)

	if err == nil {
		rs := []dao.User{}
		user.Query(&data.QueryParams{}, &rs)
		result = fmt.Sprintf("total: %d", len(rs))
	} else {
		result = err.Error()
	}

	rsp = &pb.HealthcheckResponse{
		Result: result,
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
