package sender

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mongodb/mongo-go-driver/bson"
	"log"
	"os"

	"github.com/SherClockHolmes/webpush-go"
	"lambda-push-go/core"
)

type PubSubMessage struct {
	Data []byte `json:"data"`
}

func SendNotification(ctx context.Context, m PubSubMessage) error {
	// prepare configs
	dbUrl := os.Getenv("MONGODB_URL")
	dbName := os.Getenv("DB_NAME")
	// Db connection stuff
	dbCtx := context.Background()
	dbCtx, cancel := context.WithCancel(dbCtx)
	defer cancel()

	dbCtx = context.WithValue(dbCtx, core.DbURL, dbUrl)
	db, err := core.ConfigDB(dbCtx, dbName)
	if err != nil {
		log.Fatalf("database configuration failed: %v", err)
	}

	fmt.Println("Connected to MongoDB!")

	subscriberCol := db.Collection("notificationsubscribers")
	var subscriberData core.SubscriberPayload
	err = json.Unmarshal(m.Data, &subscriberData)

	if err != nil {
		fmt.Println("json err:", err)
		return err
	}

	// Decode subscription
	s := &webpush.Subscription{}
	err = json.Unmarshal([]byte(subscriberData.PushEndpoint), s)

	if err != nil {
		fmt.Println("endpoint err:", err)
		return err
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

		return err
	}

	return nil
}
