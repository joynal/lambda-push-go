package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
	"log"

	"lambda-push-go/core"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/SherClockHolmes/webpush-go"
)

const dbUrl = "mongodb://localhost:27017"

func handler(ctx context.Context, event events.KinesisEvent) {
	// Db connection stuff
	dbCtx := context.Background()
	dbCtx, cancel := context.WithCancel(dbCtx)
	defer cancel()

	dbCtx = context.WithValue(dbCtx, core.DbURL, dbUrl)
	db, err := core.ConfigDB(dbCtx, "omnikick")
	if err != nil {
		log.Fatalf("database configuration failed: %v", err)
	}

	fmt.Println("Connected to MongoDB!")

	subscriberCol := db.Collection("notificationsubscribers")
	for _, record := range event.Records {
		_, _ = sendNotification(record.Kinesis.Data, subscriberCol, dbCtx)
	}
}

func sendNotification(record []byte, subscriberCol *mongo.Collection, dbCtx context.Context) (bool, error) {
	var subscriberData core.SubscriberPayload
	err := json.Unmarshal(record, &subscriberData)

	if err != nil {
		fmt.Println("json err:", err)
		return false, err
	}

	// Decode subscription
	s := &webpush.Subscription{}
	err = json.Unmarshal([]byte(subscriberData.PushEndpoint), s)

	if err != nil {
		fmt.Println("endpoint err:", err)
		return false, err
	}

	// Send Notification
	_, err = webpush.SendNotification([]byte(subscriberData.Data), s, &subscriberData.Options)

	// TODO: find the correct code for unsubscribe
	if err != nil {
		fmt.Println("webpush error:", err)
		_, err = subscriberCol.UpdateOne(
			dbCtx,
			bson.M{"_id": subscriberData.SubscriberID},
			bson.M{"$set": bson.M{"status": "unSubscribed"}})

		if err != nil {
			fmt.Println("db update err:", err)
		}

		return false, err
	}

	return true, nil
}

func main() {
	lambda.Start(handler)
}
