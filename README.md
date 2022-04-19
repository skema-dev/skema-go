# Skema-Go
Skema-Go is a Golang framework to simplify gRPC development by integrating various opensource components for best practice.  

## Highlight Features  
- grpcmux: grpc + http in an easy way, you can build one in less than 10 lines as below!
- database: build-in CRUD capabilities for automatic DAO(Data Access Object) generating and registration. You no long need to struggling with database! Checkout [grpc-dao sample](https://github.com/skema-dev/skema-go/tree/main/sample/grpc-dao)
- Feature Rich Configuration support


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
<br/>

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
<br/>

## About the pkg
Although it's design for grpc development, the following pkg could also be used individually.  

* /config  
  based on popular [github.com/spf13/viper](github.com/spf13/viper). It's used in grpcmux for configration handling.
  
* /logging  
  based on popular [go.uber.org/zap](go.uber.org/zap). It's used as our standard logging solution.  

* /grpcmux  
  this is the core component to integrate grpc + http in a graceful way.  
<br/>


## To Run The Code in Test
just download the repo and execute the commands below:  
```
cd test
go mod tidy
go run .
```
  
You'll see some startup information showing up. Then open another terminal and execute:  
```
# curl -X GET http://localhost:9992/api/healthcheck?msg=testuser
{"result":"health check ok"}
```
  
It's WORKING!!! Feel free to modify `server.go` to add anything you like, or use it as a start point for serious project.  
<br/>
  
  
## Use in your own project  
We do recommend using [Buf tool](https://buf.build/) or [Skemaloop gRPC Online Toolchain](https://www.skemaloop.dev) to create Protobuf file and the stubs. For testing purpose, you can use  [Skemaloop](https://www.skemaloop.dev) to generate Protobuf stubs in seconds and jump into coding.  

After the protobuf stubs are generated, you can copy the code in [/test](https://github.com/skema-dev/skema-go/tree/main/test) folder, and change the imported pb location to your online grpc stub location.  
For example, if you are using [Skemaloop](https://www.skemaloop.dev) to generate your stub, it's showing the final stub url in the `Usage` section, something like:   
```
Usage:
go get github.com/skema-repo/likezhang-public/test1/grpc-go/BB/Ttt2
```
  
Then go to the code you copied from [/test](https://github.com/skema-dev/skema-go/tree/main/test), modify the `main.go` and `server.go` files:  
```
// pb "github.com/skema-dev/skema-go/test/api/skema/test"   <===== remove or comment out this
pb "github.com/skema-repo/likezhang-public/test1/grpc-go/BB/Ttt2" // <===== use this one from skemaloop
```
  
Then, run the following commands:  
```
go mod tidy
go run .
```
  
In another terminal, use curl to verify:  
```
# curl -X GET http://localhost:9992/api/healthcheck?msg=testuser
{"result":"health check ok"}
```
  
If you have [grpcurl](https://github.com/fullstorydev/grpcurl) installed, you can also verify the grpc endpoints:  
```
# grpcurl --plaintext localhost:9991 describe
BB.Ttt2.Test is a service:
service Test {
  rpc Heathcheck ( .BB.Ttt2.HealthcheckRequest ) returns ( .BB.Ttt2.HealthcheckResponse ) {
    option (.google.api.http) = { get:"/api/healthcheck"  };
  }
  rpc Helloworld ( .BB.Ttt2.HelloRequest ) returns ( .BB.Ttt2.HelloReply ) {
    option (.google.api.http) = { post:"/api/helloworld" body:"*"  };
  }
}
grpc.reflection.v1alpha.ServerReflection is a service:
service ServerReflection {
  rpc ServerReflectionInfo ( stream .grpc.reflection.v1alpha.ServerReflectionRequest ) returns ( stream .grpc.reflection.v1alpha.ServerReflectionResponse );
}
```
  

Enjoy gRPC!