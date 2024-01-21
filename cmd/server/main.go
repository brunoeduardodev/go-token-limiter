package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"net"
	"time"

	token_collector "github.com/brunoeduardodev/go-token-limiter/contract"
	"google.golang.org/grpc"
)

var port = flag.Int("port", 50051, "The server port")

type server struct {
	token_collector.TokenCollectorServer
}

type UserBucket struct {
	tokens     float64
	lastAccess int64
}

var usersBucketPool = map[string]UserBucket{}

var tokensPerMinute = 5
var maxTokens = 5

func CreateFullUserBucket() UserBucket {
	return UserBucket{
		tokens:     float64(maxTokens),
		lastAccess: time.Now().UnixMilli(),
	}
}

func RecalculateUserBucketTokens(b *UserBucket) {
	now := time.Now().UnixMilli()
	elapsedMinutes := (float64)(now-b.lastAccess) / 1000
	tokensToInsert := elapsedMinutes * float64(tokensPerMinute)

	newTokens := math.Max(b.tokens+tokensToInsert, float64(maxTokens))
	b.tokens = newTokens
}

func (s *server) InsertToken(ctx context.Context, req *token_collector.InsertTokenRequest) (*token_collector.InsertTokenReply, error) {
	userBucket, exists := usersBucketPool[req.UserId]
	if !exists {
		userBucket := CreateFullUserBucket()
		userBucket.tokens--
		usersBucketPool[req.UserId] = userBucket
		return &token_collector.InsertTokenReply{Success: true}, nil
	}

	RecalculateUserBucketTokens(&userBucket)
	if userBucket.tokens < 1 {
		return &token_collector.InsertTokenReply{Success: false}, nil
	}

	userBucket.tokens--

	return &token_collector.InsertTokenReply{Success: true}, nil
}

func main() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))

	if err != nil {
		log.Fatalf("Could not start tcp listener: %v", err)
	}

	s := grpc.NewServer()

	token_collector.RegisterTokenCollectorServer(s, &server{})

	err = s.Serve(listener)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
