package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"lambda-push-go/core"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/primitive"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var (
	dbUrl        = "mongodb://localhost:27017"
	fcmServerKey = "config.fcmServerKey"
)

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

	// Business
	var notification core.ProcessedNotification
	record := event.Records[0]
	err = json.Unmarshal(record.Kinesis.Data, &notification)

	if err != nil {
		log.Println("json err ---->", err)
		return err
	}

	log.Println("noOfCalls:", notification.NoOfCalls)
	log.Println("totalSent:", notification.TotalSent)
	log.Println("lastId:", notification.LastID)

	query := bson.M{"siteId": notification.SiteID, "status": "subscribed"}

	if notification.Timezone != "" {
		query["timezone"] = notification.Timezone
	}

	if notification.IsFcmEnabled != "" {
		gcmAPIKey = notification.FcmServerKey
	}

	//   const webPushOptions = {
	// 	gcmAPIKey,
	// 	vapidDetails: {
	// 	  subject: 'https://omnikick.com/',
	// 	  publicKey: notification.vapidDetails.vapidPublicKeys,
	// 	  privateKey: notification.vapidDetails.vapidPrivateKeys,
	// 	},
	// 	TTL: notification.timeToLive,
	//   };

	// apply segmentation
	segmentCol := db.Collection("notificationsegments")
	if notification.SendTo.AllSubscriber == false {
		segmentQuery := bson.M{"_id": bson.M{"$in": notification.SendTo.Segments}, "isDeleted": false}
		cur, err := segmentCol.Find(dbCtx, segmentQuery)
		if err != nil {
			log.Fatal(err)
		}
		// Close the cursor once finished
		defer cur.Close(ctx)

		var segmentIds []primitive.ObjectID
		// Iterate through the cursor
		for cur.Next(ctx) {
			var elem core.Segment
			err := cur.Decode(&elem)
			if err != nil {
				log.Fatal(err)
			}

			append(segmentIds, elem.ID)
		}

		if err := cur.Err(); err != nil {
			log.Fatal(err)
		}

		query["segmentations"] = bson.M{"$in": segmentIds}
	}

	if notification.LastID != "" {
		query["_id"] = bson.M{"$gt": notification.LastID}
	}

	subscriberCol := db.Collection("notificationsubscribers")
	cur, err := subscriberCol.Find(dbCtx, query)
	if err != nil {
		log.Fatal(err)
	}
	// Close the cursor once finished
	defer cur.Close(ctx)

	// Iterate through the cursor
	for cur.Next(ctx) {
		var elem core.Subscriber
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}

	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
