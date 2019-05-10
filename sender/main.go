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

	webpush "github.com/SherClockHolmes/webpush-go"
)

const dbUrl = "mongodb://localhost:27017"

func handler(ctx context.Context, event events.KinesisEvent) error {
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

	// notification collection
	subscriberCol := db.Collection("notificationsubscribers")

	var subscriberData core.SubscriberPayload
	for _, record := range event.Records {
		_ = json.Unmarshal(record.Kinesis.Data, &subscriberData)

		// Decode subscription
		s := &webpush.Subscription{}
		_ = json.Unmarshal([]byte(subscriberData.PushEndpoint), s)

		// Send Notification
		_, err = webpush.SendNotification([]byte(subscriberData.Data), s, &webpush.Options{
			Subscriber:      subscriberData.Options.Subscriber,
			VAPIDPublicKey:  subscriberData.Options.VAPIDPublicKey,
			VAPIDPrivateKey: subscriberData.Options.VAPIDPrivateKey,
			TTL:             subscriberData.Options.TTL,
		})

		// TODO: find the correct code for unsubscribe
		if err != nil {
			log.Println(err)
			_, _ = subscriberCol.UpdateOne(dbCtx, bson.M{"_id": subscriberData.SubscriberID}, bson.M{"status": "unSubscribed"})
		}
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
