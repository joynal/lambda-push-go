package main

import (
	"flag"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
)

var (
	stream = flag.String("stream", "dev-parser", "your stream name")
	region = flag.String("region", "us-east-1", "your AWS region")
)

func main() {
	flag.Parse()

	// connect database
	// find notification & notification account
	// json string
	// send it

	s := session.New(&aws.Config{Region: aws.String(*region)})
	kc := kinesis.New(s)

	streamName := aws.String(*stream)

	putOutput, err := kc.PutRecord(&kinesis.PutRecordInput{
		Data:         []byte("hoge"),
		StreamName:   streamName,
		PartitionKey: aws.String("key1"),
	})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%v\n", putOutput)
}
