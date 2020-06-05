package controller

import (
	"context"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

const (
	// address     = "server-svc.default.svc.cluster.local:50052" //used to run over kubernetes
	address = "localhost:50052" //usedto run over local setup
)

//DialGrpcServer is a grpc-client which send a dial to grpc-server microservice
func DialGrpcServer(objectName, counter string) {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	// Contact the server and print out its response.
	name := "Component: " + objectName + " -  has running objetcs => " + counter
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	//this is a custom method which returns the message to be passed to the grpc-server
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("%s", r.Message)
}
