package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	token_collector "github.com/brunoeduardodev/go-token-limiter/contract"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var port = flag.Int("port", 50051, "The server port")
var totalResponseTimes atomic.Int64
var totalRequests atomic.Int64

func makeUserIdsList(prefix string, len int) []string {
	ids := make([]string, len)
	for i := 0; i < len; i++ {
		ids[i] = fmt.Sprintf("%s-user-id-%d", prefix, i)
	}

	return ids
}

func flooder(wg *sync.WaitGroup, client token_collector.TokenCollectorClient, flooderId, idsLen, sleepMs int) {
	userIds := makeUserIdsList(fmt.Sprintf("%d", flooderId), idsLen)

	for i := 1; i < 1000; i++ {
		userId := userIds[i%idsLen]

		now := time.Now().UnixNano()
		totalRequests.Add(1)
		_, err := client.InsertToken(context.Background(), &token_collector.InsertTokenRequest{
			UserId: userId,
		})

		totalResponseTimes.Add(time.Now().UnixNano() - now)

		if err != nil {
			log.Fatalln("Error with request", err)
		}

		time.Sleep(time.Millisecond * 10)
	}

	wg.Done()
}

func main() {
	servicePort := fmt.Sprintf("localhost:%d", *port)
	conn, err := grpc.Dial(servicePort, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalln("Could not connect to server")
	}

	client := token_collector.NewTokenCollectorClient(conn)

	var floodersCount = 10
	var wg sync.WaitGroup

	for i := 0; i < floodersCount; i++ {
		wg.Add(1)
		go flooder(&wg, client, i, 10, 1)
	}

	wg.Wait()

	result, err := client.GetBucketInformation(context.Background(), &token_collector.GetBucketInformationRequest{
		UserId: "0-user-id-0",
	})

	if err != nil {
		panic(err)
	}

	fmt.Println(result)

	fmt.Println("Cabo: ", totalRequests.Load(), totalResponseTimes.Load())

}
