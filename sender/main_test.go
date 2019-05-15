package main

import (
	"context"
	"encoding/json"
	"github.com/SherClockHolmes/webpush-go"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"lambda-push-go/core"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("parser lambda function", func() {
	var (
		ctx      context.Context
		err      error
		response bool
	)

	Context("When the notification payload is plain text", func() {
		BeforeEach(func() {
			notificationPayloadStr, _ := json.Marshal(core.NotificationPayload{
				LaunchURL: "https://joynal.github.com",
				Message: core.Message{
					Title: "Fire on ice",
					Message: "Bingo fire on ice returned",
					Language: "en",
				},
			})

			webPushOptions := webpush.Options{
				Subscriber:      "https://joynal.github.com",
				VAPIDPublicKey:  "BIUXk5gE6I8TtW5w7aoBgbA6i6o4MARqZ-fqNB7hhnzfP4hq6TIiXlqjqxn02hnuA1LBR_HmstAfTYVRBAUwAoA",
				VAPIDPrivateKey: "pLQUD5C9TrkNXQ3DoppcTUkehY9YZ5KFbbIgBwISeZo",
				TTL:             259200,
			}

			subscriberId, _ := primitive.ObjectIDFromHex("5cd664da6fc5221d583b0758")
			endpoint := `{"endpoint":"https://updates.push.services.mozilla.com/wpush/v2/gAAAAABc2vqwO0MLYQKCoj05wzNv_hZZvJRxSuiAXVJUNYp3E1y95xMpmQpSPgNvwJdlm0G9Hn6ZOA0npgvCVfR9v15DNmzR4leRs7MutPRU-_ncR-LeyxVlt2GMIyMYyWpnhJSqkM1xJtzD7lW6KwekynkV_jeuo7UzEwpRi-RMrZqK1Abglcc","keys":{"auth":"jDbPxHsnmnFvjFnyGeZ11w","p256dh":"BHS8PiMlQ2D4SDi-HzFBdlu8e3gdZWA_DS0gBEO911K40NKi-BtT9wdXS5ZybtaJ4gwABvUwk2xXUQuud0aMwkU"}}`

			data, _ := json.Marshal(core.SubscriberPayload{
				PushEndpoint: endpoint,
				Data:         string(notificationPayloadStr),
				Options:      webPushOptions,
				SubscriberID: subscriberId,
			})

			response, err = handler(ctx, core.ProcessEvent(string(data)))
		})

		It("will send notification", func() {
			Expect(err).To(BeNil())
			Expect(response).To(Equal(true))
		})
	})
})
