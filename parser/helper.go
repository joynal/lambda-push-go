package main

import "github.com/aws/aws-lambda-go/events"

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
