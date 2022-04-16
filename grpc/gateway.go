package grpc

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/skema-dev/skema-go/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type gatewayClient struct {
	svr         interface{}
	interceptor grpc.UnaryServerInterceptor
	methods     map[string]*grpc.MethodDesc
}

type gatewayConfig struct {
	Path         string
	CloudAPIPath string `yaml:"cloudapi_path"`
}

type wrapGRPCHander func(http.Handler) http.Handler

// Invoke ...
func (c *gatewayClient) Invoke(
	ctx context.Context,
	fullMethod string,
	args interface{},
	reply interface{},
	opts ...grpc.CallOption,
) error {
	pos := strings.LastIndexByte(fullMethod, '/')
	if pos < 0 {
		return fmt.Errorf("wrong method name")
	}
	logging.Infof("Full method: %s", fullMethod)
	method := fullMethod[pos+1:]
	methodHandler, ok := c.methods[method]
	if !ok {
		return fmt.Errorf("method not found: %s", method)
	}
	logging.Infof("Find method: %s", method)
	decoder := func(in interface{}) error {
		logging.Debugf("parameter in: %v", in)
		argsVal := reflect.ValueOf(args)
		rv := reflect.ValueOf(in).Elem()
		logging.Debugf("parameter rv: %v", rv)
		if !argsVal.IsZero() {
			rv.Set(argsVal.Elem())
		}
		return nil
	}
	// incoming ctx to outgoing ctx
	md, ok := metadata.FromOutgoingContext(ctx)
	if ok {
		ctx = metadata.NewIncomingContext(ctx, md)
	}
	rsp, err := methodHandler.Handler(c.svr, ctx, decoder, c.interceptor)
	if err != nil {
		return err
	}
	rv := reflect.ValueOf(reply).Elem()
	rspVal := reflect.ValueOf(rsp)
	if !rspVal.IsZero() {
		rv.Set(rspVal.Elem())
	}
	return nil
}

// NewStream ...
func (c *gatewayClient) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string,
	opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, fmt.Errorf("un implemented")
}
