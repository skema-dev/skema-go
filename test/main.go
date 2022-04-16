package main

import (
	"github.com/skema-dev/skema-go/grpc"
	"github.com/skema-dev/skema-go/logging"
	pb "github.com/skema-dev/skema-go/test/api/skema/test"
)

func main() {
	grpcSrv := grpc.NewServer()

	srvImp := NewServer()
	pb.RegisterTestServer(grpcSrv, srvImp)

	ctx, mux, conn := grpcSrv.GetGatewayInfo(srvImp, &pb.Test_ServiceDesc)
	pb.RegisterTestHandlerClient(ctx, mux, pb.NewTestClient(conn))

	if err := grpcSrv.Serve(); err != nil {
		logging.Fatalf("Serve error %v", err.Error())
	}
}
