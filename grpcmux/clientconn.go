package grpcmux

import (
	"sync"

	"github.com/skema-dev/skema-go/config"
	"github.com/skema-dev/skema-go/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ClientConn struct {
	conns *sync.Map
	opts  []grpc.DialOption
}

// GetConn create a new client conn
func GetConn() grpc.ClientConnInterface {
	localConfig := LoadLocalConfig()
	return NewClientConnWithConfig(localConfig)
}

// NewClientConnWithConfig create a new client conn with config
func (cc *ClientConn) NewClientConnWithConfig(config *config.Config) grpc.ClientConnInterface {
	conn, err := grpc.Dial(config.GetString("client.address"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logging.Fatalf("Did not connect: %v", err)
	}
	cc.conns.Store(,conn)
	return conn
}
