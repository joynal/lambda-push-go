package main

import (
	"flag"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
)

var (
	stream = flag.String("stream", "test-parser", "your stream name")
	region = flag.String("region", "us-east-1", "your AWS region")
)

func main() {
	flag.Parse()

	s := session.New(&aws.Config{Region: aws.String(*region)})
	kc := kinesis.New(s)

	streamName := aws.String(*stream)

	// put 10 records using PutRecords API
	entries := make([]*kinesis.PutRecordsRequestEntry, 10)
	for i := 0; i < len(entries); i++ {
		entries[i] = &kinesis.PutRecordsRequestEntry{
			Data:         []byte(fmt.Sprintf("test-%d", i)),
			PartitionKey: aws.String("key2"),
		}
	}
	fmt.Printf("%v\n", entries)
	putsOutput, err := kc.PutRecords(&kinesis.PutRecordsInput{
		Records:    entries,
		StreamName: streamName,
	})
	if err != nil {
		panic(err)
	}
	// putsOutput has Records, and its shard id and sequence number.
	fmt.Printf("%v\n", putsOutput)
}
