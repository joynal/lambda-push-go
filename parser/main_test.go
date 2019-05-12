package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"lambda-push-go/core"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("parser lambda function", func() {
	var (
		response core.ProcessedNotification
		ctx      context.Context
		event    events.KinesisEvent
		err      error
	)

	AfterEach(func() {
		response = core.ProcessedNotification{}
		ctx = context.TODO()
		event = events.KinesisEvent{}
	})

	Context("When the notification payload is plain text", func() {
		BeforeEach(func() {
			kc := &mockKinesisClient{}
			lc := &mockLambdaClient{}
			var records []events.KinesisEventRecord
			records = append(records, events.KinesisEventRecord{
				Kinesis: events.KinesisRecord{
					Data: []byte("Test string"),
				},
			})
			event = events.KinesisEvent{
				Records: records,
			}
			parserLambda := handler(kc, lc)
			response, err = parserLambda(ctx, event)
		})

		It("Fails", func() {
			Expect(err).To(MatchError("invalid character 'T' looking for beginning of value"))
			Expect(response).To(Equal(core.ProcessedNotification{}))
		})
	})

	Context("When the notification payload is correct json", func() {
		BeforeEach(func() {
			kc := &mockKinesisClient{}
			lc := &mockLambdaClient{}
			notificationID, _ := primitive.ObjectIDFromHex("5c82455744cd0f069b35daa6")
			siteID, _ := primitive.ObjectIDFromHex("5c82424627ff1506951b7fbb")

			notificationStr, _ := json.Marshal(core.ProcessedNotification{
				ID:         notificationID,
				SiteID:     siteID,
				TimeToLive: 259200,
				LaunchURL:  "https://joynal.github.io",
				Message: core.Message{
					Title:    "Fire on ice",
					Message:  "Bingo fire on ice returned",
					Language: "en",
				},
				Browser: []core.Browser{
					{
						BrowserName: "chrome",
						IconURL:     "https://cdn.omnikick.com/assets/img/Logo_Smile_blue.png",
						Vibration:   false,
					},
					{
						BrowserName: "firefox",
						IconURL:     "https://cdn.omnikick.com/assets/img/Logo_Smile_blue.png",
						Vibration:   false,
					},
				},
				TotalSent:     0,
				NoOfCalls: 0,
				SendTo:        core.SendTo{
					AllSubscriber: true,
				},
				IsAtLocalTime: false,

				// notification account data
				VapidDetails: core.VapidDetails{
					VapidPrivateKeys: "",
					VapidPublicKeys: "",
				},
			})
			var records []events.KinesisEventRecord
			records = append(records, events.KinesisEventRecord{
				Kinesis: events.KinesisRecord{
					Data: []byte(notificationStr),
				},
			})
			event = events.KinesisEvent{
				Records: records,
			}
			parserLambda := handler(kc, lc)
			response, err = parserLambda(ctx, event)
		})

		It("should move to next", func() {
			Expect(err).To(BeNil())
			Expect(response.NoOfCalls).To(Equal(1))
			Expect(response.TotalSent).To(Equal(10))
			Expect(response.LastID.Hex()).To(Equal("5cd664da6fc5221d583b0761"))
		})
	})
})
