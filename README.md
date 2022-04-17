# Skema-Go
Skema-Go is a Golang framework to simplify gRPC development by integrating various opensource components for best practice.

## grpcmux: gRPC + http service in 10 lines
Talk is cheap. First, let's see how we can create a grpc server with http enabled in just 10 lines.
```
func main() {
	grpcSrv := grpcmux.NewServer()                # create our grpc server with some addional properties
	pb.RegisterTestServer(grpcSrv, NewServer())   # register service on grpc server

	ctx, mux, conn := grpcSrv.GetGatewayInfo()                      # we need some information for http binding
	pb.RegisterTestHandlerClient(ctx, mux, pb.NewTestClient(conn))  # register for http gateway

	if err := grpcSrv.Serve(); err != nil {       # start server
		logging.Fatalf("Serve error %v", err.Error())
	}
}
```
* NewServer() is the grpc service to be implenented. You can check [/test](https://github.com/skema-dev/skema-go/tree/main/test) for details.  

It's much simplified and prettier than the [standard grpc + gateway solution](https://grpc-ecosystem.github.io/grpc-gateway/docs/tutorials/adding_annotations/)  

So, where is the port defined? Just some encapsulation?  Let's check the following features:

## IaC and Configuration
Everything is defined in [/test/grpc.yaml](https://github.com/skema-dev/skema-go/tree/main/test/grpc.yaml), or you can also load from your remote endpoint like etcd or consul. Let's use the local grpc.yaml for example:  
```
port: 9991     # for grpc service
http:
  port: 9992   # for http service
  gateway:
    # path: "/test1/"
    path: "/"  # this is the routing prefix if you need. You can force to append an extra path before all standard URLs

logging:
  level: debug # info | debug
  encoding: console # console | json
  output: "./log/default.log"
```
  
Pretty Clear. We can define the grpc listening port and http port in the config, as well as some other features. If you've used Django, this is pretty much theh same idea.  

## About the pkg
* /config  
  based on popular [github.com/spf13/viper](github.com/spf13/viper). It's used in grpcmux for configration handling.
  
* /logging  
  based on popular [go.uber.org/zap](go.uber.org/zap). It's used as our standard logging solution.  

* /grpcmux  
  this is the core component to integrate grpc + http in a graceful way.  

