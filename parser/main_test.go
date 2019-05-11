package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/lambda"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("parser lambda function", func() {
	var (
		response lambda.InvokeOutput
		ctx  context.Context
		event events.KinesisEvent
		err      error
	)

	JustBeforeEach(func() {
		var records []events.KinesisEventRecord
		records = append(records, events.KinesisEventRecord{
			Kinesis: events.KinesisRecord{
				Data: []byte("Test string"),
			},
		})
		event = events.KinesisEvent{
			Records: records,
		}
		response, err = handler(ctx, event)
	})

	AfterEach(func() {
		response = lambda.InvokeOutput{}
		ctx = context.TODO()
		event = events.KinesisEvent{}
	})

	Context("When the notification payload is plain text", func() {
		It("Fails", func() {
			Expect(err).To(MatchError("invalid character 'T' looking for beginning of value"))
			Expect(response).To(Equal(lambda.InvokeOutput{}))
		})
	})

	Context("When the notification payload is correct json", func() {
		BeforeEach(func() {
			var records []events.KinesisEventRecord
			records = append(records, events.KinesisEventRecord{
				Kinesis: events.KinesisRecord{
					Data: []byte(""),
				},
			})
			event = events.KinesisEvent{
				Records: records,
			}
			response, err = handler(ctx, event)
		})
	})
})
