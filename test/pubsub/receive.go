package main

import (
	"cloud.google.com/go/pubsub"
	"context"
	"fmt"
	"log"
	"os"
	"sync"
)

func main() {
	ctx := context.Background()

	client, err := pubsub.NewClient(ctx, os.Getenv("GOOGLE_CLOUD_PROJECT"))
	if err != nil {
		log.Fatal(err)
	}

	var mu sync.Mutex
	sub := client.Subscription("notification")
	cctx, _ := context.WithCancel(ctx)
	err = sub.Receive(cctx, func(ctx context.Context, msg *pubsub.Message) {
		msg.Ack()
		fmt.Printf("Got message: %q\n", string(msg.Data))
		mu.Lock()
		defer mu.Unlock()
	})
	if err != nil {
		fmt.Println(err)
	}
}

