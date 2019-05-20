package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/SherClockHolmes/webpush-go"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/kinesis/kinesisiface"
	lambdaSdk "github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/google/uuid"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/mongodb/mongo-go-driver/mongo/options"
	"lambda-push-go/core"
	"log"
	"os"
	"strconv"
	"time"
)

func handler(kc kinesisiface.KinesisAPI, lc lambdaiface.LambdaAPI) func(context.Context, events.KinesisEvent) (core.ProcessedNotification, error) {
	return func (ctx context.Context, event events.KinesisEvent) (core.ProcessedNotification, error) {
		// prepare configs
		dbUrl := os.Getenv("MONGODB_URL")
		dbName := os.Getenv("DB_NAME")
		batchSize, _ := strconv.Atoi(os.Getenv("PARSER_BATCH_SIZE"))
		queryBatchSize, _ := strconv.Atoi(os.Getenv("QUERY_BATCH_SIZE"))

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

		// process notification json object to struct
		var notification core.ProcessedNotification
		record := event.Records[0]
		err = json.Unmarshal(record.Kinesis.Data, &notification)

		if err != nil {
			return core.ProcessedNotification{}, err
		}

		fmt.Println("noOfCalls:", notification.NoOfCalls)
		fmt.Println("totalSent:", notification.TotalSent)
		fmt.Println("lastId:", notification.LastID)

		// increase no of calls
		notification.NoOfCalls += 1

		// Lets prepare subscriber query
		query := bson.M{
			"_id": bson.M{"$gt": notification.LastID},
			"siteId": notification.SiteID,
			"status": "subscribed",
		}

		// if notification have timezone
		if notification.Timezone != "" {
			query["timezone"] = notification.Timezone
		}

		// apply segmentation to subscriber query
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

		// query options, limit size
		opts := options.Find()
		opts.SetLimit(int64(queryBatchSize))

		subscriberCol := db.Collection("notificationsubscribers")
		var subscribers []*kinesis.PutRecordsRequestEntry

		// find subscribers and
		cur, err := subscriberCol.Find(dbCtx, query, opts)
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
				Data:         string(notificationPayloadStr),
				Options:      webPushOptions,
				SubscriberID: elem.ID,
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
		streamName := aws.String(os.Getenv("SENDER_STREAM_NAME"))
		if len(subscribers) > 0 {
			_, err := kc.PutRecords(&kinesis.PutRecordsInput{
				Records:    subscribers,
				StreamName: streamName,
			})
			if err != nil {
				fmt.Println("put records error:", err)
			}
		}

		// update total sent
		notification.TotalSent += len(subscribers)

		// finish the recursion & update notification
		if len(subscribers) < batchSize {
			updateQuery := bson.M{"updatedAt": time.Now()}
			updateQuery["totalSent"] = notification.TotalSent

			if notification.IsAtLocalTime == false {
				updateQuery["isProcessed"] = "done"
			}

			notificationCol := db.Collection("notifications")
			_, _ = notificationCol.UpdateOne(dbCtx, bson.M{"_id": notification.ID}, bson.M{"$set": updateQuery})

			return notification, nil
		}

		// invoke recursive way
		payload, _ := json.Marshal(notification)
		result, err := lc.Invoke(&lambdaSdk.InvokeInput{
			FunctionName:   aws.String(lambdacontext.FunctionName),
			Qualifier:      aws.String(lambdacontext.FunctionVersion),
			InvocationType: aws.String("Event"),
			Payload:        payload,
		})

		if err != nil {
			fmt.Println("invoke error:", err)
		}

		fmt.Println("invoke result:", string(result.Payload))

		if err == nil && *result.StatusCode == 200 {
			return notification, nil
		}

		return core.ProcessedNotification{}, nil
	}
}

func main() {
	s, _ := session.NewSession(&aws.Config{Region: aws.String("us-east-1")})
	kc := kinesis.New(s)
	lc := lambdaSdk.New(s)

	lambda.Start(handler(kc, lc))
}
