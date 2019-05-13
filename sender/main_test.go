package main

import (
	"context"
	"lambda-push-go/core"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("parser lambda function", func() {
	var (
		ctx      context.Context
		err      error
		response string
	)

	Context("When the notification payload is plain text", func() {
		BeforeEach(func() {
			response, err = handler(ctx, core.ProcessEvent("Test string"))
		})

		It("Fails", func() {
			Expect(err).To(MatchError("invalid character 'T' looking for beginning of value"))
			Expect(response).To(Equal(""))
		})
	})
})
