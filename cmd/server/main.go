package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	token_collector "github.com/brunoeduardodev/go-token-limiter/contract"
	token_bucket "github.com/brunoeduardodev/go-token-limiter/internal"
	"google.golang.org/grpc"
)

var port = flag.Int("port", 50051, "The server port")

type server struct {
	tokenMachine token_bucket.TokenMachine
	token_collector.TokenCollectorServer
}

func (s *server) InsertToken(ctx context.Context, req *token_collector.InsertTokenRequest) (*token_collector.InsertTokenReply, error) {
	result := s.tokenMachine.InsertToken(req.UserId)
	if result {
		return &token_collector.InsertTokenReply{Success: true}, nil
	}

	return &token_collector.InsertTokenReply{Success: false}, nil
}

func main() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))

	if err != nil {
		log.Fatalf("Could not start tcp listener: %v", err)
	}

	s := grpc.NewServer()

	token_collector.RegisterTokenCollectorServer(s, &server{
		tokenMachine: *token_bucket.MakeTokenMachine(5, 5),
	})

	err = s.Serve(listener)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
