package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/lambdacontext"

	"lambda-push-go/core"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/primitive"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	lambdaSdk "github.com/aws/aws-sdk-go/service/lambda"
	"github.com/google/uuid"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var (
	dbUrl     = "mongodb://localhost:27017"
	gcmAPIKey = "config.fcmServerKey"
	batchSize = 500
	stream    = flag.String("stream", "test-parser", "your stream name")
	region    = flag.String("region", "us-east-1", "your AWS region")
)

func handler(ctx context.Context, event events.KinesisEvent) error {
	flag.Parse()

	s := session.New(&aws.Config{Region: aws.String(*region)})
	kc := kinesis.New(s)
	lambdaClient := lambdaSdk.New(s)

	streamName := aws.String(*stream)

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
	notificationCol := db.Collection("notifications")

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

			segmentIds = append(segmentIds, elem.ID)
		}

		if err := cur.Err(); err != nil {
			log.Fatal(err)
		}

		query["segmentations"] = bson.M{"$in": segmentIds}
	}

	if primitive.String(notification.LastID) == "" {
		query["_id"] = bson.M{"$gt": notification.LastID}
	}

	if notification.IsFcmEnabled {
		gcmAPIKey = notification.FcmServerKey
	}

	webPushOptions := core.WebPushOptions{
		GcmAPIKey: gcmAPIKey,
		VapidDetails: core.VAPIDOptions{
			Subject:    "https://omnikick.com/",
			PublicKey:  notification.VapidDetails.VapidPublicKeys,
			PrivateKey: notification.VapidDetails.VapidPrivateKeys,
		},
		TTL: notification.TimeToLive,
	}

	notificationPayload := core.NotificationPayload{
		ID:        notification.ID,
		LaunchURL: notification.LaunchURL,
		Message:   notification.Message,
		Browser:   notification.Browser,
		HideRules: notification.HideRules,
		Actions:   notification.Actions,
	}

	subscriberCol := db.Collection("notificationsubscribers")
	var subscribers []*kinesis.PutRecordsRequestEntry
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

		id := uuid.New()

		processed, _ := json.Marshal(core.SubscriberPayload{
			PushEndpoint: elem.PushEndpoint,
			Data:         notificationPayload,
			Options:      webPushOptions,
			SubscriberId: elem.ID,
		})

		subscribers = append(subscribers, &kinesis.PutRecordsRequestEntry{
			Data:         processed,
			PartitionKey: aws.String(id.String()),
		})

		notification.LastID = elem.ID
	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	// send to kinesis
	if len(subscribers) > 0 {
		putsOutput, err := kc.PutRecords(&kinesis.PutRecordsInput{
			Records:    subscribers,
			StreamName: streamName,
		})
		if err != nil {
			panic(err)
		}
		// putsOutput has Records, and its shard id and sequence number.
		fmt.Printf("%v\n", putsOutput)
	}

	notification.TotalSent += len(subscribers)

	// finish the recursion
	if len(subscribers) < batchSize {
		updateQuery := bson.M{"updatedAt": time.Now()}
		// TODO: fix total sent conflict later
		updateQuery["totalSent"] = notification.TotalSent
		if notification.IsAtLocalTime == false {
			updateQuery["isProcessed"] = "done"
		}

		notificationCol.UpdateOne(dbCtx, bson.M{"_id": notification.ID}, updateQuery)

		return nil
	}

	// invoke recursive way

	notification.NoOfCalls += 1
	payload, _ := json.Marshal(notification)

	result, err := lambdaClient.Invoke(&lambdaSdk.InvokeInput{
		FunctionName:   aws.String(lambdacontext.FunctionName),
		Qualifier:      aws.String(lambdacontext.FunctionVersion),
		InvocationType: aws.String("Event"),
		Payload:        payload,
	})
	if err != nil {
		fmt.Println("invoke:", result)
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
