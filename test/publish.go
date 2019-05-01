package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"lambda-push-go/core"

	"github.com/mongodb/mongo-go-driver/bson"
)

var (
	// stream = flag.String("stream", "dev-parser", "your stream name")
	// region = flag.String("region", "us-east-1", "your AWS region")
	dbUrl = "mongodb://localhost:27017"
)

func main() {
	flag.Parse()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ctx = context.WithValue(ctx, core.DbURL, dbUrl)
	db, err := core.ConfigDB(ctx, "pushservice")
	if err != nil {
		log.Fatalf("database configuration failed: %v", err)
	}

	fmt.Println("Connected to MongoDB!")

	var notification core.Notification
	var notificationAccount core.NotificationAccount

	err = db.Collection("notifications").FindOne(ctx, bson.D{}).Decode(&notification)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("notification: ", notification)

	err = db.Collection("notificationaccounts").FindOne(ctx, bson.D{}).Decode(&notificationAccount)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("notificationAccount: ", notificationAccount)

	// find notification & notification account
	// json string
	// send it

	// s := session.New(&aws.Config{Region: aws.String(*region)})
	// kc := kinesis.New(s)

	// streamName := aws.String(*stream)

	// putOutput, err := kc.PutRecord(&kinesis.PutRecordInput{
	// 	Data:         []byte("hoge"),
	// 	StreamName:   streamName,
	// 	PartitionKey: aws.String("key1"),
	// })
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Printf("%v\n", putOutput)
}
