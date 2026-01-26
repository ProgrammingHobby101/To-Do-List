//source code: @"Step 3 — main.go" , https://chatgpt.com/c/69767414-0c18-8329-bf21-8a28445ee16d

package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var (
	dbClient  *dynamodb.Client
	tableName = os.Getenv("TABLE_NAME")
)

// ✅ Your struct
type User struct {
	UserID string `json:"userId" dynamodbav:"userId"`
	Name   string `json:"name" dynamodbav:"name"`
	Email  string `json:"email" dynamodbav:"email"`
}

// Runs once per container (cold start)
func init() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatal("unable to load AWS SDK config:", err)
	}

	dbClient = dynamodb.NewFromConfig(cfg)
}

func handler(ctx context.Context, req events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {

	switch req.RequestContext.HTTP.Method {

	case "POST":
		return createUser(ctx, req)

	case "GET":
		return getUser(ctx, req)

	default:
		return response(405, map[string]string{
			"error": "method not allowed",
		})
	}
}

// ### 🔹 Create User (POST)
func createUser(ctx context.Context, req events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {

	var user User

	if err := json.Unmarshal([]byte(req.Body), &user); err != nil {
		return response(400, map[string]string{"error": "invalid json"})
	}

	if user.UserID == "" || user.Name == "" || user.Email == "" {
		return response(400, map[string]string{"error": "missing fields"})
	}

	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		return response(500, map[string]string{"error": "marshal failed"})
	}

	_, err = dbClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	})

	if err != nil {
		log.Println("PutItem error:", err)
		return response(500, map[string]string{"error": "dynamodb error"})
	}

	return response(201, map[string]string{
		"message": "user created",
	})
}
func getUser(ctx context.Context, req events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {

	userID := req.QueryStringParameters["userId"]

	if userID == "" {
		return response(400, map[string]string{"error": "userId required"})
	}

	result, err := dbClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"userId": &types.AttributeValueMemberS{Value: userID},
		},
	})

	if err != nil {
		log.Println("GetItem error:", err)
		return response(500, map[string]string{"error": "dynamodb error"})
	}

	if result.Item == nil {
		return response(404, map[string]string{"error": "not found"})
	}

	var user User
	err = attributevalue.UnmarshalMap(result.Item, &user)
	if err != nil {
		return response(500, map[string]string{"error": "unmarshal error"})
	}

	return response(200, user)
}
func response(code int, body any) (events.LambdaFunctionURLResponse, error) {
	jsonBody, _ := json.Marshal(body)

	return events.LambdaFunctionURLResponse{
		StatusCode: code,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(jsonBody),
	}, nil
}
