package main

import (
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/kinesis/kinesisiface"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
)

type mockKinesisClient struct {
	kinesisiface.KinesisAPI
}

func (m *mockKinesisClient) PutRecords(*kinesis.PutRecordsInput) (*kinesis.PutRecordsOutput, error) {
	return &kinesis.PutRecordsOutput{}, nil
}

type mockLambdaClient struct {
	lambdaiface.LambdaAPI
}

func (m *mockLambdaClient) Invoke(*lambda.InvokeInput) (*lambda.InvokeOutput, error) {
	return &lambda.InvokeOutput{}, nil
}
