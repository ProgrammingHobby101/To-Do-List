// source: main.go@"Lambda Code (Using events)" ; https://chatgpt.com/c/6975761d-148c-8327-85fd-15c01dc752c4

package main

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type User struct {
	UserID   string `json:"userId" dynamodbav:"userId"`
	Name     string `json:"name" dynamodbav:"name"`
	Email    string `json:"email" dynamodbav:"email"`
	Password string `json:"password" dynamodbav:"password"`
}

var tableName = "To-Do-List-Users"

func handler(
	ctx context.Context,
	request events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {

	// Parse JSON body
	var user User
	err := json.Unmarshal([]byte(request.Body), &user)
	if err != nil {
		return clientError(400, "Invalid JSON body")
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return serverError(err)
	}

	client := dynamodb.NewFromConfig(cfg)

	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		return serverError(err)
	}

	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &tableName,
		Item:      item,
	})
	if err != nil {
		return serverError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 201,
		Body:       "User created successfully",
	}, nil
}

func clientError(code int, msg string) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: code,
		Body:       msg,
	}, nil
}

func serverError(err error) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: 500,
		Body:       err.Error(),
	}, nil
}
func main() {
	// if val := os.Getenv("TABLE_NAME"); val != "" {
	// 	tableName = val
	// }
	lambda.Start(handler)
}
