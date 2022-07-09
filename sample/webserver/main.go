package main

import (
	_ "embed"

	"github.com/skema-dev/skema-go/config"
	"github.com/skema-dev/skema-go/grpcmux"
	"github.com/skema-dev/skema-go/logging"
	pb "github.com/skema-dev/skema-go/sample/api/skema/test"
)

//go:embed grpc.yaml
var yamlConfig []byte

func main() {
	grpcSrv := grpcmux.NewServerWithConfig(
		config.NewConfigWithString(string(yamlConfig)),
	)

	pb.RegisterTestServer(grpcSrv, NewServer())

	// for http gateway only.
	ctx, mux, conn := grpcSrv.GetGatewayInfo()
	pb.RegisterTestHandlerClient(ctx, mux, pb.NewTestClient(conn))

	//grpcSrv.EnableStaticContent("/", "./static")
	//grpcSrv.EnableStaticContent("/script", "./static/script")

	if err := grpcSrv.Serve(); err != nil {
		logging.Fatalf("Serve error %v", err.Error())
	}
}
