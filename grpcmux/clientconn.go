package grpcmux

import (
	"context"
	"fmt"
	"sync"

	"github.com/skema-dev/skema-go/config"
	"github.com/skema-dev/skema-go/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ClientConn struct {
	mux    sync.Mutex
	conns  []*grpc.ClientConn
	ctx    context.Context
	cancel context.CancelFunc
}

// GetConn create a new client conn
func GetConn() (grpc.ClientConnInterface, error) {
	localConfig := LoadLocalConfig()
	ctx, cancelFunc := context.WithCancel(context.Background())
	conns := &ClientConn{
		ctx:    ctx,
		cancel: cancelFunc,
	}
	conn, err := conns.NewClientConnWithConfig(localConfig)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// NewClientConnWithConfig create a new client conn with config
func (cc *ClientConn) NewClientConnWithConfig(config *config.Config) (grpc.ClientConnInterface, error) {
	if config.GetString("client.address") == "" {
		logging.Fatalf("Can not get client address in config file, please check!")
	}
	cc.mux.Lock()
	cc.mux.Unlock()
	conn, err := grpc.DialContext(cc.ctx, config.GetString("client.address"),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logging.Fatalf("Did not connect: %v", err)
		defer cc.cancel()
		return nil, fmt.Errorf("connect error! ")
	}
	cc.conns = append(cc.conns, conn)
	return conn, nil
}

// Close the connections
func (cc *ClientConn) Close() error {
	defer cc.cancel()
	cc.mux.Lock()
	defer cc.mux.Unlock()
	for _, conn := range cc.conns {
		conn.Close()
	}
	return nil
}
