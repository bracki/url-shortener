package main

import (
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	log "github.com/sirupsen/logrus"
)

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.WithField("body", request.Body).Info("Received request")
	s := request.PathParameters["short_url"]
	log.WithField("short_url", s).Info("Got short URL")

	sess := session.Must(session.NewSession())
	dynamodbclient := dynamodb.New(sess)
	result, err := dynamodbclient.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(os.Getenv("DYNAMO_DB_TABLE")),
		Key: map[string]*dynamodb.AttributeValue{
			"short_url": {S: aws.String(s)},
		},
	})
	if err != nil {
		log.WithField("error", err).Info("Couldn't get data from DynamoDB")
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: err.Error()}, nil
	}
	if result.Item == nil {
		return events.APIGatewayProxyResponse{StatusCode: 404, Body: "nix"}, nil
	}
	loc := result.Item["url"]

	return events.APIGatewayProxyResponse{StatusCode: 302, Headers: map[string]string{"Location": *loc.S}}, nil
}

func main() {
	lambda.Start(Handler)
}
