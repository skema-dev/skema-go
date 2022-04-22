package server

import (
	"context"
	"fmt"
	"log"

	"grpc-dao/internal/model"

	"github.com/skema-dev/skema-go/data"
	"github.com/skema-dev/skema-go/logging"
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

	user := data.Manager().GetDAO(&model.User{})
	err = user.Upsert(&model.User{
		UUID: uuid.New().String(),
		Name: req.Msg,
	}, nil, nil)

	if err == nil {
		rs := []model.User{}
		user.Query(&data.QueryParams{}, &rs)
		result = fmt.Sprintf("total: %d", len(rs))

		_ = model.Address{}

		if len(rs) > 3 {
			err = user.Delete("name like 'user%'")
			if err != nil {
				logging.Errorf(err.Error())
			}
		}

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
