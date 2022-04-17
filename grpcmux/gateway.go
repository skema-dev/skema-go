package grpcmux

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
)

type gatewayClient struct {
	connection grpc.ClientConnInterface
}

func (c *gatewayClient) Invoke(ctx context.Context, fullMethod string, args interface{}, reply interface{},
	opts ...grpc.CallOption) error {
	if c.connection == nil {
		return nil
	}
	return c.connection.Invoke(ctx, fullMethod, args, reply, opts...)
}

func (c *gatewayClient) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string,
	opts ...grpc.CallOption) (grpc.ClientStream, error) {
	if c.connection == nil {
		return nil, fmt.Errorf("client connection not created yet")
	}
	return c.connection.NewStream(ctx, desc, method, opts...)
}
