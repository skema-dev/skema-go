package main

import (
	"github.com/skema-dev/skema-go/grpc"
	"github.com/skema-dev/skema-go/logging"
	pb "github.com/skema-dev/test/api/skema/test"
)

func main() {
	grpcSrv := grpc.NewServer()

	srvImp := NewServer()
	pb.RegisterTest111Server(grpcSrv, srvImp)

	ctx, mux, conn := grpcSrv.GetGatewayInfo(srvImp, &pb.Test111_ServiceDesc)
	pb.RegisterTest111HandlerClient(ctx, mux, pb.NewTest111Client(conn))

	if err := grpcSrv.Serve(); err != nil {
		logging.Fatalf("Serve error %v", err.Error())
	}
}
