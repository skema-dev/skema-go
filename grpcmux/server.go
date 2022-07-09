package grpcmux

import (
	"context"
	"flag"
	"fmt"
	"github.com/skema-dev/skema-go/config"
	"github.com/skema-dev/skema-go/logging"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

var (
	configSearchPaths = []string{"./grpc.yaml", "./config/grpc.yaml", "/config/grpc.yaml", "./grpc.yml", "./config/grpc.yml", "/config/grpc.yml"}
)

type grpcServer struct {
	conf   *config.Config
	server *grpc.Server
	port   int

	httpPort         int
	httpMux          *http.ServeMux
	gatewayMux       *runtime.ServeMux
	gatewayRoutePath string

	clientConn *gatewayClient

	ctx        context.Context
	cancelFunc context.CancelFunc
}

func NewServer(opts ...grpc.ServerOption) *grpcServer {
	localConfig := LoadLocalConfig()
	return NewServerWithConfig(localConfig, opts...)
}

// Add additional setup besides standard grpc.NewServer
func NewServerWithConfig(conf *config.Config, opts ...grpc.ServerOption) *grpcServer {
	ctx, cancelFunc := context.WithCancel(context.Background())

	port := conf.GetInt("port")
	httpPort := conf.GetInt("http.port")
	if port == httpPort {
		logging.Fatalw("http port is the same as grpc port", "grpc port", port, "http port", httpPort)
	}
	if port == 0 {
		logging.Fatalf("please specify port in config")
	}
	logging.Infow("service port", "gprc", port, "http", httpPort)

	if !validateHttpConfig(conf) {
		logging.Fatalf("duplicated url path found. please fix the grpc config file")
	}

	// connect to grpc port
	conn, err := grpc.DialContext(
		context.Background(),
		"localhost"+fmt.Sprintf(":%d", port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logging.Errorf("Failed to create connection for localhost:%d: %s", port, err.Error())
		cancelFunc()
		return nil
	}

	serverMux := runtime.NewServeMux(
		runtime.WithUnescapingMode(runtime.UnescapingModeAllExceptReserved),
	)

	gatewayPathPrefix := "/"
	if httpPort > 0 {
		gatewayPathPrefix = conf.GetString("http.gateway.path")
		if gatewayPathPrefix == "" {
			gatewayPathPrefix = "/"
		} else {
			if !strings.HasSuffix(gatewayPathPrefix, "/") {
				gatewayPathPrefix += "/"
			}
		}
	}

	initComponents(conf)

	srv := grpc.NewServer(
		opts...,
	)

	return &grpcServer{
		conf:             conf,
		server:           srv,
		httpMux:          http.NewServeMux(),
		gatewayMux:       serverMux,
		gatewayRoutePath: gatewayPathPrefix,
		ctx:              ctx,
		cancelFunc:       cancelFunc,
		port:             port,
		httpPort:         httpPort,
		clientConn:       &gatewayClient{connection: conn},
	}
}

// load config from local file
func LoadLocalConfig() *config.Config {
	// look for local config file
	var path string
	flag.StringVar(&path, "config", "", "path for grpc server config")
	flag.Parse()

	if path == "" {
		for _, p := range configSearchPaths {
			if _, err := os.Stat(p); !os.IsNotExist(err) {
				path = p
				break
			}
		}
	}
	if path == "" {
		msg := fmt.Sprintf(
			"config path not found, please put grpc.yaml in one of %v or set --config=xxx.yaml",
			configSearchPaths,
		)
		panic(msg)
	}
	logging.Infof("using local config from %s", path)

	return config.NewConfigWithFile(path)
}

func validateHttpConfig(conf *config.Config) bool {
	gatewayPath := conf.GetString("http.gateway.path", "")
	staticPath := conf.GetString("http.static.path", "")
	swaggerPath := conf.GetString("http.swagger.path", "")

	values := []string{gatewayPath, staticPath, swaggerPath}
	for i := range values {
		if values[i] == "" {
			continue
		}
		j := i + 1
		for j < len(values) {
			if values[i] == values[j] {
				logging.Errorf("duplicated url prefix path: %s, %s\n", values[i], values[j])
				return false
			}
			j += 1
		}
	}
	return true
}

// Requeired for grpc service registration
func (g *grpcServer) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	g.server.RegisterService(desc, impl)
}

func (g *grpcServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := g.gatewayRoutePath
	if len(path) > 1 {
		// only do the trick when we have a route prefix other than "/"
		if strings.HasSuffix(path, "/") {
			// remove last "/", e.g. "/abc/" => "/abc", so the endpoint would match
			path = strings.TrimSuffix(path, "/")
		}
		r.URL.Path = strings.TrimPrefix(r.URL.Path, path)
		r.RequestURI = strings.TrimPrefix(r.RequestURI, path)
	}
	logging.Infof("server http %s\n", r.URL.Path)
	g.gatewayMux.ServeHTTP(w, r)
}

// Start serving grpc and http server
func (g *grpcServer) Serve() error {
	reflection.Register(g.server)

	g.httpMux.Handle(g.gatewayRoutePath, g)
	logging.Infof("grpc-gateway path: %s", g.gatewayRoutePath)

	if g.conf.GetString("http.static.path", "") != "" {
		staticPath := g.conf.GetString("http.static.path")
		if !strings.HasSuffix(staticPath, "/") {
			staticPath += "/"
		}
		staticFilepath := g.conf.GetString("http.static.filepath")
		staticHandler := http.FileServer(http.Dir(staticFilepath))

		g.httpMux.Handle(staticPath, http.StripPrefix(staticPath, staticHandler))

		logging.Infof("static content(%s) path: %s", staticFilepath, staticPath)
	}

	if g.conf.GetString("http.swagger.path", "") != "" {
		swaggerPath := g.conf.GetString("http.swagger.path")
		swaggerPath = strings.TrimSuffix(swaggerPath, "/")
		swaggerFilepath := g.conf.GetString("http.swagger.filepath")
		swaggerHandler, openapiHandler := g.getSwaggerHandler(swaggerFilepath)

		g.httpMux.Handle(swaggerPath, swaggerHandler)
		g.httpMux.Handle(fmt.Sprintf("%s/openapi", swaggerPath), openapiHandler)

		logging.Infof("swagger path: %s", swaggerPath)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", g.port))
	if err != nil {
		logging.Fatalf("failed listening on port %d", g.port)
	}
	defer g.cancelFunc()

	if g.httpPort > 0 {
		go func() {
			if err := http.ListenAndServe(fmt.Sprintf(":%d", g.httpPort), g.httpMux); err != nil {
				logging.Fatalf(err.Error())
			}
		}()
	}

	return g.server.Serve(lis)
}

// GetGatewayInfo 返回Http网关相关信息
func (g *grpcServer) GetGatewayInfo() (context.Context, *runtime.ServeMux, grpc.ClientConnInterface) {
	return g.ctx, g.gatewayMux, g.clientConn
}

func (g *grpcServer) getSwaggerHandler(openapiDescFilepath string) (http.HandlerFunc, http.HandlerFunc) {
	openapiHandler := func(w http.ResponseWriter, r *http.Request) {
		if content, err := ioutil.ReadFile(openapiDescFilepath); err == nil {
			fmt.Fprint(w, string(content))
		} else {
			log.Fatal(err)
		}
	}

	swaggerHandler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, swaggerTpl)
	}

	return swaggerHandler, openapiHandler
}
