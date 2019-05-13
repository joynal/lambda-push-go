package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go/service/kinesis/kinesisiface"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"lambda-push-go/core"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("parser lambda function", func() {
	var (
		response core.ProcessedNotification
		ctx      context.Context
		err      error
		kc kinesisiface.KinesisAPI
		lc lambdaiface.LambdaAPI
		notification core.ProcessedNotification
	)

	BeforeEach(func() {
		kc = &mockKinesisClient{}
		lc = &mockLambdaClient{}
		notificationID, _ := primitive.ObjectIDFromHex("5c82455744cd0f069b35daa6")
		siteID, _ := primitive.ObjectIDFromHex("5c82424627ff1506951b7fbb")
		notification = core.ProcessedNotification{
			ID:         notificationID,
			SiteID:     siteID,
			TotalSent:     0,
			NoOfCalls: 0,
			SendTo:        core.SendTo{
				AllSubscriber: true,
			},
			IsAtLocalTime: false,
		}
	})

	AfterEach(func() {
		response = core.ProcessedNotification{}
		ctx = context.TODO()
		kc = &mockKinesisClient{}
		lc = &mockLambdaClient{}
		notification = core.ProcessedNotification{}
	})

	Context("When the notification payload is plain text", func() {
		BeforeEach(func() {
			parserLambda := handler(kc, lc)
			response, err = parserLambda(ctx, processEvent("Test string"))
		})

		It("Fails", func() {
			Expect(err).To(MatchError("invalid character 'T' looking for beginning of value"))
			Expect(response).To(Equal(core.ProcessedNotification{}))
		})
	})

	Context("Testing first page", func() {
		BeforeEach(func() {
			notificationStr, _ := json.Marshal(notification)
			parserLambda := handler(kc, lc)
			response, err = parserLambda(ctx, processEvent(string(notificationStr)))
		})

		It("should move to next", func() {
			Expect(err).To(BeNil())
			Expect(response.NoOfCalls).To(Equal(1))
			Expect(response.TotalSent).To(Equal(10))
			Expect(response.LastID.Hex()).To(Equal("5cd664da6fc5221d583b0761"))
		})
	})

	Context("Testing second page", func() {
		BeforeEach(func() {
			lastId, _ := primitive.ObjectIDFromHex("5cd664da6fc5221d583b0761")
			notification.NoOfCalls = 1
			notification.TotalSent = 10
			notification.LastID = lastId
			notificationStr, _ := json.Marshal(notification)
			parserLambda := handler(kc, lc)
			response, err = parserLambda(ctx, processEvent(string(notificationStr)))
		})

		It("should move to next", func() {
			Expect(err).To(BeNil())
			Expect(response.NoOfCalls).To(Equal(2))
			Expect(response.TotalSent).To(Equal(20))
			Expect(response.LastID.Hex()).To(Equal("5cd664da6fc5221d583b076b"))
		})
	})

	Context("Testing third page", func() {
		BeforeEach(func() {
			lastId, _ := primitive.ObjectIDFromHex("5cd664da6fc5221d583b076b")
			notification.NoOfCalls = 2
			notification.TotalSent = 20
			notification.LastID = lastId
			notificationStr, _ := json.Marshal(notification)
			parserLambda := handler(kc, lc)
			response, err = parserLambda(ctx, processEvent(string(notificationStr)))
		})

		It("should move to next", func() {
			Expect(err).To(BeNil())
			Expect(response.NoOfCalls).To(Equal(3))
			Expect(response.TotalSent).To(Equal(30))
			Expect(response.LastID.Hex()).To(Equal("5cd664da6fc5221d583b0775"))
		})
	})

	Context("Testing fourth page", func() {
		BeforeEach(func() {
			lastId, _ := primitive.ObjectIDFromHex("5cd664da6fc5221d583b0775")
			notification.NoOfCalls = 3
			notification.TotalSent = 30
			notification.LastID = lastId
			notificationStr, _ := json.Marshal(notification)
			parserLambda := handler(kc, lc)
			response, err = parserLambda(ctx, processEvent(string(notificationStr)))
		})

		It("should move to next", func() {
			Expect(err).To(BeNil())
			Expect(response.NoOfCalls).To(Equal(4))
			Expect(response.TotalSent).To(Equal(40))
			Expect(response.LastID.Hex()).To(Equal("5cd664da6fc5221d583b077f"))
		})
	})

	Context("Testing fifth page", func() {
		BeforeEach(func() {
			lastId, _ := primitive.ObjectIDFromHex("5cd664da6fc5221d583b077f")
			notification.NoOfCalls = 4
			notification.TotalSent = 40
			notification.LastID = lastId
			notificationStr, _ := json.Marshal(notification)
			parserLambda := handler(kc, lc)
			response, err = parserLambda(ctx, processEvent(string(notificationStr)))
		})

		It("should move to next", func() {
			Expect(err).To(BeNil())
			Expect(response.NoOfCalls).To(Equal(5))
			Expect(response.TotalSent).To(Equal(50))
			Expect(response.LastID.Hex()).To(Equal("5cd664da6fc5221d583b0789"))
		})
	})

	Context("Testing sixth page", func() {
		BeforeEach(func() {
			lastId, _ := primitive.ObjectIDFromHex("5cd664da6fc5221d583b0789")
			notification.NoOfCalls = 5
			notification.TotalSent = 50
			notification.LastID = lastId
			notificationStr, _ := json.Marshal(notification)
			parserLambda := handler(kc, lc)
			response, err = parserLambda(ctx, processEvent(string(notificationStr)))
		})

		It("should move to next", func() {
			Expect(err).To(BeNil())
			Expect(response.NoOfCalls).To(Equal(6))
			Expect(response.TotalSent).To(Equal(55))
			Expect(response.LastID.Hex()).To(Equal("5cd664da6fc5221d583b078e"))
		})
	})
})
