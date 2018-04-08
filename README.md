# Steps

```
curl -XPOST -d '{"url":"https://superluminar.io"}' https://s5fnxrrwz5.execute-api.us-east-1.amazonaws.com/dev/create-url
```

```
serverless create -t aws-go-dep -p url-shortener

dep ensure -add github.com/sirupsen/logrus
dep ensure -add github.com/aws/aws-sdk-go/aws
dep ensure -add github.com/kelseyhightower/envconfig
```

```
rm -fr world
```

```
vi serverless.yml
```

Rename hello to create-url

events:
  - http:
      path: create-url
      method: post

Edit create-url/main.go

```
package main

import (
    "fmt"
    
    "github.com/aws/aws-lambda-go/events"
    "github.com/aws/aws-lambda-go/lambda"
)


func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Println("Received body: ", request.Body)

	return events.APIGatewayProxyResponse{Body: request.Body, StatusCode: 200}, nil
}

func main() {
	lambda.Start(Handler)
}
```

Add log message
```
	log.WithField("body", request.Body).Info("Received request")
```


read JSON data, form data is too hard.. ;(

```
var data map[string]string
err := json.Unmarshal([]byte(request.Body), &data)
if err != nil {
	log.WithField("error", err).Info("Error while reading request")
	return events.APIGatewayProxyResponse{Body: request.Body, StatusCode: 500}, nil
}
if url, ok := data["url"]; !ok {
	return events.APIGatewayProxyResponse{Body: "no url", StatusCode: 400}, nil
}
```

paste shorten func

```
// Shorten shortens a URL and will return an error if the URL does not validate.
// The implementation is a bit naive but good enough for a show case.
func Shorten(u string) (string, error) {
        if _, err := url.ParseRequestURI(u); err != nil {
                return "", err
        }
        hash := fnv.New64a()
        hash.Write([]byte(u))
        return strconv.FormatUint(hash.Sum64(), 36), nil
}
```

```
	s, err := Shorten(u)
	if err != nil {
		log.WithField("error", err).Error("Malformed URL")
		return events.APIGatewayProxyResponse{Body: "Malformed URL", StatusCode: 400}, nil
	}
	body := fmt.Sprintf("%s/%s", request.Headers["Host"], s)
	return events.APIGatewayProxyResponse{Body: body, StatusCode: 201}, nil
```

Add get-url

```
  get-url:
    handler: bin/get-url
    events:
      - http:
          path: /{short_url} 
          method: get
          request:
            parameters:
              paths:
                short_url: true
```

Copy create-url to get-url

```

package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	log "github.com/sirupsen/logrus"
)

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.WithField("body", request.Body).Info("Received request")
	s := request.PathParameters["short_url"]
	log.WithField("short_url", s).Info("Got short URL")
	return events.APIGatewayProxyResponse{StatusCode: 302, Headers: map[string]string{"Location": "https://superluminar.io"}}, nil
}

func main() {
	lambda.Start(Handler)
}
```


IAM and CFN

```
  iamRoleStatements:
    -  Effect: "Allow"
       Action:
         - "dynamodb:PutItem"
         - "dynamodb:GetItem"
       Resource:
         Fn::GetAtt:
           - DynamoDBTable
           - Arn


resources:
  Resources:
    DynamoDBTable:
      Type: AWS::DynamoDB::Table
      Properties:
        TableName: url-shortener 
        KeySchema:
          - AttributeName: "short_url"
            KeyType: "HASH"
        ProvisionedThroughput:
          ReadCapacityUnits: "1"
          WriteCapacityUnits: "1"
        AttributeDefinitions:
          - AttributeName: "short_url"
            AttributeType: "S"

```

Dynamo code

```
	sess := session.Must(session.NewSession())
	dynamodbclient := dynamodb.New(sess)
	_, err = dynamodbclient.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(os.Getenv("DYNAMO_DB_TABLE")),
		Item: map[string]*dynamodb.AttributeValue{
			"short_url": &dynamodb.AttributeValue{S: aws.String(s)},
			"url":       &dynamodb.AttributeValue{S: aws.String(u)},
		}})
	if err != nil {
		log.WithField("error", err).Error("Couldn't save URL")
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 500}, nil
	}
```
