// Copyright © 2017 David King <davidgwking@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package sqsjanitor

import (
	"errors"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type QueueDetails struct {
	QueueURL     string
	MessageCount int
}

var sqsClient *sqs.SQS

func getSqsClient() *sqs.SQS {
	if sqsClient != nil {
		return sqsClient
	}

	session, err := session.NewSession()
	if err != nil {
		log.Fatal(err)
	}

	sqsClient := sqs.New(session)

	return sqsClient
}

func getQueueList(client *sqs.SQS) (*sqs.ListQueuesOutput, error) {
	input := &sqs.ListQueuesInput{}
	output, err := client.ListQueues(input)
	if err != nil {
		return nil, err
	}

	return output, err
}

func getQueueAttributes(client *sqs.SQS, queueUrl string) (*sqs.GetQueueAttributesOutput, error) {
	var approxNumMessages = "All"
	input := &sqs.GetQueueAttributesInput{
		AttributeNames: []*string{&approxNumMessages},
		QueueUrl:       &queueUrl,
	}
	output, err := client.GetQueueAttributes(input)
	if err != nil {
		return nil, err
	}

	return output, err
}

func GetQueueDetails(maxWorkers int) (*QueueListModel, error) {
	client := getSqsClient()

	queues, err := getQueueList(client)
	if err != nil {
		return nil, err
	}

	numQueues := len(queues.QueueUrls)
	queueURLs := make(chan string, numQueues)
	outputs := make(chan *QueueDetails, numQueues)
	errs := make(chan error)
	done := make(chan bool)

	for i := 0; i < maxWorkers; i++ {
		go QueueAttributesWorker(queueURLs, outputs, errs, done)
	}

	for _, queueURL := range queues.QueueUrls {
		queueURLs <- *queueURL
	}
	close(queueURLs)

	for i := 0; i < maxWorkers; i++ {
		<-done
	}

	details := make(QueueListModel, 0, numQueues)
	for i := 0; i < numQueues; i++ {
		details = append(details, <-outputs)
	}

	return &details, nil
}

func PurgeQueue(queueURL string) error {
	client := getSqsClient()

	input := &sqs.PurgeQueueInput{}
	input.SetQueueUrl(queueURL)

	_, err := client.PurgeQueue(input)
	if err != nil {
		return err
	}

	return nil
}

func QueueAttributesWorker(queueURLs <-chan string, outputs chan<- *QueueDetails, errs chan<- error, done chan<- bool) {
	client := getSqsClient()

	for queueURL := range queueURLs {
		output, err := getQueueAttributes(client, queueURL)
		if err != nil {
			errs <- err
		}

		if numMessagesS, ok := output.Attributes["ApproximateNumberOfMessages"]; !ok {
			errs <- errors.New("ApproximateNumberOfMessages")
		} else {
			numMessages, err := strconv.Atoi(*numMessagesS)
			if err != nil {
				errs <- err
			} else {
				outputs <- &QueueDetails{queueURL, numMessages}
			}
		}
	}
	done <- true
}
