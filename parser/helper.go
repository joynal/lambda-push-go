package main

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"lambda-push-go/core"
)

func processEvent(str string) events.KinesisEvent {
	var records []events.KinesisEventRecord
	records = append(records, events.KinesisEventRecord{
		Kinesis: events.KinesisRecord{
			Data: []byte(str),
		},
	})
	return events.KinesisEvent{
		Records: records,
	}
}

func getNotificaitonStr(notification core.ProcessedNotification, lastIdStr string, noOfCalls int, totalSent int) string {
	lastId, _ := primitive.ObjectIDFromHex(lastIdStr)
	notification.NoOfCalls = noOfCalls
	notification.TotalSent = totalSent
	notification.LastID = lastId
	notificationStr, _ := json.Marshal(notification)
	return  string(notificationStr)
}
