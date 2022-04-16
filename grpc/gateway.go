package grpc

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type gatewayClient struct {
	svr         interface{}
	interceptor grpc.UnaryServerInterceptor
	methods     map[string]*grpc.MethodDesc
}

type wrapGRPCHander func(http.Handler) http.Handler

func (g *gatewayClient) Invoke(
	ctx context.Context,
	fullMethod string,
	args interface{},
	reply interface{},
	opts ...grpc.CallOption,
) error {
	pos := strings.LastIndexByte(fullMethod, '/')
	if pos < 0 {
		return fmt.Errorf("invalid method name: %s", fullMethod)
	}

	method := fullMethod[pos+1:]
	methodHandler, ok := g.methods[method]
	if !ok {
		return fmt.Errorf("method [%s] doesn't exist", method)
	}

	decoder := func(in interface{}) error {
		argsVal := reflect.ValueOf(args)
		receiver := reflect.ValueOf(in).Elem()
		if !argsVal.IsZero() {
			receiver.Set(argsVal.Elem())
		}
		return nil
	}

	// incoming ctx to outgoing ctx
	md, ok := metadata.FromOutgoingContext(ctx)
	if ok {
		ctx = metadata.NewIncomingContext(ctx, md)
	}

	rsp, err := methodHandler.Handler(g.svr, ctx, decoder, g.interceptor)
	if err != nil {
		return err
	}
	receiver := reflect.ValueOf(reply).Elem()
	rspVal := reflect.ValueOf(rsp)
	if !rspVal.IsZero() {
		receiver.Set(rspVal.Elem())
	}

	return nil
}

func (g *gatewayClient) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string,
	opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, fmt.Errorf("not implemented")
}
