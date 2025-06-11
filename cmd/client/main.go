package main

import (
	"context"
	"github.com/jialuohu/curlyraft/internal/proto/gateway"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
)

func main() {
	conn, err := grpc.NewClient("unix:///tmp/raft.sock",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to build up connection between client and socket file")
	}
	defer conn.Close()
	client := gateway.NewRaftGatewayClient(conn)
	res, err := client.AppendCommand(
		context.Background(),
		&gateway.AppendCommandRequest{Command: []byte("hello world")})

	if err != nil {
		log.Fatalf("Failed to get response from grpc server.")
	}

	if res.GetApplied() {
		log.Printf("Got Applied!!!!!")
	} else {
		log.Printf("Faild")
	}

}
