package main

import (
	"context"
	"fmt"

	_ "embed"

	"github.com/skema-dev/skema-go/grpcmux"
	"github.com/skema-dev/skema-go/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func main() {
	grpcSrv := grpcmux.NewServer(
		grpc.ChainUnaryInterceptor(Interceptor1(), Interceptor2(), Interceptor3()),
	)

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

func Interceptor3() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp interface{}, err error) {
		_, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, fmt.Errorf("couldn't parse incoming context metadata")
		}
		logging.Infof(" interceptoer 333333")
		h, err := handler(ctx, req)
		return h, err
	}
}
