package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mongodb/mongo-go-driver/bson"
	"log"

	"lambda-push-go/core"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/SherClockHolmes/webpush-go"
)

const dbUrl = "mongodb://localhost:27017"

func handler(ctx context.Context, event events.KinesisEvent) (bool, error) {
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
	var subscriberData core.SubscriberPayload
	for _, record := range event.Records {
		err = json.Unmarshal(record.Kinesis.Data, &subscriberData)

		if err != nil {
			fmt.Println("json err:", err)
		}

		// Decode subscription
		s := &webpush.Subscription{}
		_ = json.Unmarshal([]byte(subscriberData.PushEndpoint), s)

		if err != nil {
			fmt.Println("endpoint err:", err)
		}

		// Send Notification
		_, err := webpush.SendNotification([]byte(subscriberData.Data), s, &subscriberData.Options)

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
		}
	}

	return true, nil
}

func main() {
	lambda.Start(handler)
}
