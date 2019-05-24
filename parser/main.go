package main

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"fmt"
	"github.com/SherClockHolmes/webpush-go"
	"github.com/joho/godotenv"
	"github.com/mongodb/mongo-go-driver/bson"
	"lambda-push-go/core"
	"log"
	"os"
	"sync"
	"time"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// prepare configs
	dbUrl := os.Getenv("MONGODB_URL")
	dbName := os.Getenv("DB_NAME")
	parserTopic := os.Getenv("PARSER_TOPIC")
	senderTopic := os.Getenv("SENDER_TOPIC")

	// Db connection stuff
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ctx = context.WithValue(ctx, core.DbURL, dbUrl)
	db, err := core.ConfigDB(ctx, dbName)
	if err != nil {
		log.Fatalf("database configuration failed: %v", err)
	}

	fmt.Println("Connected to MongoDB!")

	// process notification json object to struct
	client, err := pubsub.NewClient(ctx, os.Getenv("GOOGLE_CLOUD_PROJECT"))
	if err != nil {
		log.Fatal(err)
	}

	var mu sync.Mutex
	var notification core.ProcessedNotification
	sub := client.Subscription(parserTopic)
	cctx, _ := context.WithCancel(ctx)
	err = sub.Receive(cctx, func(ctx context.Context, msg *pubsub.Message) {
		msg.Ack()
		err = json.Unmarshal(msg.Data, &notification)
		if err != nil {
			fmt.Println(err)
		}
		mu.Lock()
		defer mu.Unlock()
	})
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("notification: %v\n", notification)

	// Lets prepare subscriber query
	query := bson.M{
		"siteId": notification.SiteID,
		"status": "subscribed",
	}

	// if notification have timezone
	if notification.Timezone != "" {
		query["timezone"] = notification.Timezone
	}

	webPushOptions := webpush.Options{
		Subscriber:      "https://omnikick.com/",
		VAPIDPublicKey:  notification.VapidDetails.VapidPublicKeys,
		VAPIDPrivateKey: notification.VapidDetails.VapidPrivateKeys,
		TTL:             notification.TimeToLive,
	}

	notificationPayload := core.NotificationPayload{
		ID:        notification.ID,
		LaunchURL: notification.LaunchURL,
		Message:   notification.Message,
		Browser:   notification.Browser,
		HideRules: notification.HideRules,
		Actions:   notification.Actions,
	}

	notificationPayloadStr, _ := json.Marshal(notificationPayload)

	subscriberCol := db.Collection("notificationsubscribers")

	// find subscribers and
	cur, err := subscriberCol.Find(ctx, query)
	if err != nil {
		log.Fatal(err)
	}
	// Close the cursor once finished
	defer cur.Close(ctx)

	// Iterate through the cursor
	topic := client.Topic(senderTopic)
	for cur.Next(ctx) {
		var elem core.Subscriber
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}

		processed, _ := json.Marshal(core.SubscriberPayload{
			PushEndpoint: elem.PushEndpoint,
			Data:         string(notificationPayloadStr),
			Options:      webPushOptions,
			SubscriberID: elem.ID,
		})

		result := topic.Publish(ctx, &pubsub.Message{
			Data: []byte(processed),
		})

		id, err := result.Get(ctx)
		if err != nil {
			fmt.Println(err)
		}

		notification.TotalSent++

		fmt.Printf("Sent msg ID: %v\n", id)
	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	// finish the recursion & update notification
	updateQuery := bson.M{"updatedAt": time.Now()}
	updateQuery["totalSent"] = notification.TotalSent

	if notification.IsAtLocalTime == false {
		updateQuery["isProcessed"] = "done"
	}

	notificationCol := db.Collection("notifications")
	_, _ = notificationCol.UpdateOne(ctx, bson.M{"_id": notification.ID}, bson.M{"$set": updateQuery})
}
