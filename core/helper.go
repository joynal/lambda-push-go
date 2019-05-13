package core

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

func ProcessEvent(str string) events.KinesisEvent {
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

func GetNotificaitonStr(notification ProcessedNotification, lastIdStr string, noOfCalls int, totalSent int) string {
	lastId, _ := primitive.ObjectIDFromHex(lastIdStr)
	notification.NoOfCalls = noOfCalls
	notification.TotalSent = totalSent
	notification.LastID = lastId
	notificationStr, _ := json.Marshal(notification)
	return  string(notificationStr)
}
