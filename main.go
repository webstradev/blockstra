package main

import (
	"context"
	"log"
	"time"

	"github.com/webstradev/blockstra/node"
	"github.com/webstradev/blockstra/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const vers = "blockstra-0.1"

func main() {
	makeNode(":3000", []string{})
	time.Sleep(50 * time.Millisecond)
	makeNode(":4000", []string{":3000"})
	time.Sleep(100 * time.Millisecond)
	makeNode(":5000", []string{":4000"})

	select {}
}

func makeNode(listenAddr string, bootstrapNodes []string) *node.Node {
	loggerConfig := zap.NewDevelopmentConfig()
	loggerConfig.EncoderConfig.TimeKey = "timestamp"

	zap, err := loggerConfig.Build()
	if err != nil {
		log.Fatal(err)
	}
	n := node.New(vers, listenAddr, zap.Sugar(), bootstrapNodes)
	go n.Start()
	return n
}

func makeTransaction() {
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	client, err := grpc.Dial(":3000", opts...)
	if err != nil {
		log.Fatal(err)
	}

	c := proto.NewNodeClient(client)

	version := &proto.Version{
		Version:    "blockstra-0.1",
		Height:     1,
		ListenAddr: ":4000",
	}

	_, err = c.Handshake(context.Background(), version)
	if err != nil {
		log.Fatal(err)
	}
}
