package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/sqs"
)

const (
	noop int = iota
	stop
	start

	RUNNING = "RUNNING"
	STOPPED = "STOPPED"
)

type ECSContainerOverride struct {
	Name string `json:"name"`
}

type ECSOverrides struct {
	ContainerOverrides []ECSContainerOverride `json:"containerOverrides"`
}

type ECSEventDetail struct {
	DesiredStatus     string       `json:"desiredStatus"`
	LastStatus        string       `json:"lastStatus"`
	Overrides         ECSOverrides `json:"overrides"`
	TaskDefinitionArn string       `json:"taskDefinitionArn"`
}

type ECSEvent struct {
	Time    time.Time      `json:"time"`
	Detail  ECSEventDetail `json:"detail"`
	Status  string         `json:"-"`
	Command []string       `json:"-"`
}

func ECSEvents() <-chan *ECSEvent {
	out := make(chan *ECSEvent)
	config := &aws.Config{
		Region: aws.String(os.Getenv("AWS_DEFAULT_REGION")),
	}
	svc := sqs.New(session.Must(session.NewSession(config)))
	queueName := os.Getenv("SQS_QUEUE")
	url, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{QueueName: &queueName})
	if err != nil {
		log.Fatal(err)
	}
	ecsCli := ecs.New(session.Must(session.NewSession(config)))
	ticker := time.NewTicker(time.Second)
	go func(out chan<- *ECSEvent) {
		for range ticker.C {
			output, err := svc.ReceiveMessage(&sqs.ReceiveMessageInput{QueueUrl: url.QueueUrl})
			if err != nil {
				log.Println(err)
			}
			for _, msg := range output.Messages {
				e := new(ECSEvent)
				err = json.Unmarshal([]byte(*msg.Body), e)
				if err != nil {
					log.Println(err)
					continue
				}
				e.Status = strings.Title(strings.ToLower(status(e)))
				cont, err := getTaskDetails(ecsCli, e.Detail.TaskDefinitionArn)
				if err != nil {
					log.Println(err)
					continue
				}
				if len(cont.ContainerDefinitions) > 0 {
					e.Command = sliceConv(cont.ContainerDefinitions[0].Command)
					out <- e
				}
				svc.DeleteMessage(&sqs.DeleteMessageInput{
					QueueUrl:      url.QueueUrl,
					ReceiptHandle: msg.ReceiptHandle,
				})
			}
		}
	}(out)
	return out
}

func getTaskDetails(cli *ecs.ECS, arn string) (*ecs.TaskDefinition, error) {
	input := &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: &arn,
	}
	result, err := cli.DescribeTaskDefinition(input)
	if err != nil {
		return nil, err
	}
	return result.TaskDefinition, nil
}

func status(e *ECSEvent) string {
	status := STOPPED
	if e.Detail.DesiredStatus == STOPPED {
		status = STOPPED
	}
	if e.Detail.DesiredStatus == RUNNING {
		status = RUNNING
	}
	return status
}
