package main

import (
	"context"
	"log"
	"time"

	"github.com/webstradev/blockstra/crypto"
	"github.com/webstradev/blockstra/node"
	"github.com/webstradev/blockstra/proto"
	"github.com/webstradev/blockstra/util"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const vers = "blockstra-0.1"

func main() {
	cfg := node.ServerConfig{
		Version:    vers,
		ListenAddr: ":3000",
		PrivateKey: crypto.MustGeneratePrivateKey(),
	}
	makeNode(cfg, []string{})

	time.Sleep(50 * time.Millisecond)
	cfg = node.ServerConfig{
		Version:    vers,
		ListenAddr: ":4000",
	}
	makeNode(cfg, []string{":3000"})

	time.Sleep(50 * time.Millisecond)
	cfg = node.ServerConfig{
		Version:    vers,
		ListenAddr: ":5000",
	}
	makeNode(cfg, []string{":4000"})

	for {
		time.Sleep(2 * time.Second)
		makeTransaction()
	}
}

func makeNode(cfg node.ServerConfig, bootstrapNodes []string) *node.Node {
	loggerConfig := zap.NewDevelopmentConfig()
	loggerConfig.EncoderConfig.TimeKey = "timestamp"

	zap, err := loggerConfig.Build()
	if err != nil {
		log.Fatal(err)
	}
	n := node.New(cfg, zap.Sugar(), bootstrapNodes)
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

	privKey := crypto.MustGeneratePrivateKey()
	tx := &proto.Transaction{
		Version: 1,
		Inputs: []*proto.TxInput{
			{
				PrevTxHash:   util.RandomHash(),
				PrevOutIndex: 0,
				PublicKey:    privKey.Public().Bytes(),
			},
		},
		Outputs: []*proto.TxOutput{
			{
				Amount:  99,
				Address: privKey.Public().Address().Bytes(),
			},
		},
	}

	_, err = c.HandleTransaction(context.Background(), tx)
	if err != nil {
		log.Fatal(err)
	}
}
