package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"lambda-push-go/core"

	"github.com/google/uuid"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/mongodb/mongo-go-driver/bson"
)

const dbUrl = "mongodb://localhost:27017"
const stream = "test-parser"
const region = "us-east-1"

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ctx = context.WithValue(ctx, core.DbURL, dbUrl)
	db, err := core.ConfigDB(ctx, "omnikick")
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

	err = db.Collection("notificationaccounts").FindOne(ctx, bson.D{}).Decode(&notificationAccount)

	if err != nil {
		log.Fatal(err)
	}

	processed, _ := json.Marshal(core.ProcessedNotification{
		ID:            notification.ID,
		SiteID:        notification.SiteID,
		TimeToLive:    notification.TimeToLive,
		LaunchURL:     notification.LaunchURL,
		Message:       notification.Messages[0],
		Browser:       notification.Browsers,
		HideRules:     notification.HideRules,
		TotalSent:     notification.TotalSent,
		SendTo:        notification.SendTo,
		IsAtLocalTime: false,

		// notification account data
		IsFcmEnabled: notificationAccount.IsFcmEnabled,
		FcmSenderId:  notificationAccount.FcmSenderId,
		FcmServerKey: notificationAccount.FcmServerKey,
		VapidDetails: notificationAccount.VapidDetails,
	})

	s, _ := session.NewSession(&aws.Config{Region: aws.String(region)})
	kc := kinesis.New(s)

	streamName := aws.String(stream)
	id := uuid.New()

	putOutput, err := kc.PutRecord(&kinesis.PutRecordInput{
		Data:         processed,
		StreamName:   streamName,
		PartitionKey: aws.String(id.String()),
	})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%v\n", putOutput)
}
