package main

import (
	"context"
	"fmt"

	_ "embed"

	"grpc-dao/cmd/server"

	"grpc-dao/internal/dao"

	"github.com/skema-dev/skema-go/database"
	"github.com/skema-dev/skema-go/grpcmux"
	"github.com/skema-dev/skema-go/logging"
	pb "github.com/skema-dev/skema-go/sample/api/skema/test"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func main() {
	grpcSrv := grpcmux.NewServer(
		grpc.ChainUnaryInterceptor(Interceptor1(), Interceptor2()),
	)

	dao := database.Manager().GetDAO("", dao.User{})
	fmt.Printf("name: %v\n", dao)

	pb.RegisterTestServer(grpcSrv, server.NewServer())

	// for http gateway only.
	ctx, mux, conn := grpcSrv.GetGatewayInfo()
	pb.RegisterTestHandlerClient(ctx, mux, pb.NewTestClient(conn))

	if err := grpcSrv.Serve(); err != nil {
		logging.Fatalf("Serve error %v", err.Error())
	}
}

func Interceptor1() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp interface{}, err error) {
		_, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, fmt.Errorf("couldn't parse incoming context metadata")
		}
		logging.Infof("first interceptoer 11111")
		h, err := handler(ctx, req)
		return h, err
	}
}

func Interceptor2() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp interface{}, err error) {
		_, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, fmt.Errorf("couldn't parse incoming context metadata")
		}
		logging.Infof(" interceptoer 222222")
		h, err := handler(ctx, req)
		return h, err
	}
}
