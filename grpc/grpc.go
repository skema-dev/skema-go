package grpc

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/skema-dev/skema-go/config"
	"github.com/skema-dev/skema-go/logging"
	"go.uber.org/automaxprocs/maxprocs"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	configPaths = []string{"./grpc.yaml", "./config/grpc.yaml", "/config/grpc.yaml"}
)

type grpcServer struct {
	serverConfig *config.Config

	srv      *grpc.Server
	port     int
	httpPort int

	httpMux    *http.ServeMux
	gatewayMux *runtime.ServeMux

	interceptor grpc.UnaryServerInterceptor

	ctx context.Context
}

func NewServer(opt ...grpc.ServerOption) *grpcServer {
	var path string
	flag.StringVar(&path, "config", "", "server config path")
	flag.Parse()

	if path == "" {
		for _, p := range configPaths {
			if _, err := os.Stat(p); !os.IsNotExist(err) {
				path = p
				break
			}
		}
	}

	if path == "" {
		logging.Panicf("please put grpc.yaml in one of %v or set --config=xxx.yaml", configPaths)
	}
	serverCfg := config.NewConfigWithFile(path)

	srv := grpc.NewServer(opt...)

	return &grpcServer{
		serverConfig: serverCfg,
		srv:          srv,
		port:         serverCfg.GetInt("port"),
		httpPort:     serverCfg.GetInt("http_port"),
		httpMux:      http.NewServeMux(),
		gatewayMux:   runtime.NewServeMux(),
		ctx:          context.Background(),
	}
}

// RegisterService ...
func (gs *grpcServer) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	gs.srv.RegisterService(desc, impl)
}

// Serve ...
func (gs *grpcServer) Serve() error {
	cfg := gs.serverConfig

	reflection.Register(gs.srv)

	maxprocs.Set(maxprocs.Logger(func(s string, args ...interface{}) {
		logging.Debugf(s, args...)
	}))

	gatewayPath := cfg.GetString("gateway.path")
	if len(gatewayPath) > 0 {
		logging.Infof("gateway path: %s", gatewayPath)
		gs.httpMux.Handle(gatewayPath, gs.gatewayMux)
	}

	// handle the situation when grpc/http use the same port
	if gs.httpPort > 0 && gs.port == gs.httpPort {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ProtoMajor == 2 && strings.HasPrefix(
				r.Header.Get("Content-Type"), "application/grpc") {
				gs.srv.ServeHTTP(w, r)
			} else {
				gs.httpMux.ServeHTTP(w, r)
			}
		})
		h2s := &http2.Server{}
		h1s := &http.Server{
			Addr:    fmt.Sprintf(":%d", gs.httpPort),
			Handler: h2c.NewHandler(handler, h2s),
		}

		if err := h1s.ListenAndServe(); err != nil {
			logging.Fatalf(err.Error())
		}
		return nil
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", gs.port))
	if err != nil {
		return err
	}

	// running http gateway
	go func() {
		if gs.httpPort > 0 {
			logging.Infof("http  started at %d", gs.httpPort)
			if err := http.ListenAndServe(fmt.Sprintf(":%d", gs.httpPort), gs.httpMux); err != nil {
				logging.Fatalf(err.Error())
			}
		}
	}()

	return gs.srv.Serve(lis)
}

// GetGatewayInfo 返回Http网关相关信息
func (gs *grpcServer) GetGatewayInfo(imp interface{},
	desc *grpc.ServiceDesc) (context.Context, *runtime.ServeMux, grpc.ClientConnInterface) {
	methods := make(map[string]*grpc.MethodDesc)

	for i := range desc.Methods {
		m := &desc.Methods[i]
		methods[m.MethodName] = m
	}
	cc := &gatewayClient{svr: imp, methods: methods, interceptor: gs.interceptor}
	return gs.ctx, gs.gatewayMux, cc
}
