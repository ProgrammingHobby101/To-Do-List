package main

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var (
	dbClient  *dynamodb.Client
	tableName = "To-Do-List-Users"
)

//////////////////////
// STRUCTS
//////////////////////

type User struct {
	UserID   string `json:"userId" dynamodbav:"userId"`
	Name     string `json:"name" dynamodbav:"name"`
	Email    string `json:"email" dynamodbav:"email"`
	Password string `json:"password" dynamodbav:"password"`
}

type LoginUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

//////////////////////
// INIT
//////////////////////

func init() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatal("unable to load AWS SDK config:", err)
	}

	dbClient = dynamodb.NewFromConfig(cfg)
}

//////////////////////
// HANDLER
//////////////////////

func handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {

	log.Println("request:", req.RequestContext.HTTP.Method, req.RequestContext.HTTP.Path)

	method := req.RequestContext.HTTP.Method
	path := req.RequestContext.HTTP.Path

	switch method + " " + path {
	case "POST /api/to-do-list/mypost/users":
		return createUser(ctx, req)

	case "HEAD /api/to-do-list/mypost/health":
		return handleHello(ctx, req)

	case "OPTIONS /api/to-do-list/mypost/users":
		return response(200, map[string]string{"message": "ok"})

	default:
		return response(405, map[string]string{"error": "method not allowed"})
	}
}

//////////////////////
// HEALTH
//////////////////////

func handleHello(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return response(204, nil)
}

//////////////////////
// CREATE USER
//////////////////////

func createUser(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {

	var user User

	if err := json.Unmarshal([]byte(req.Body), &user); err != nil {
		log.Println("createUser unmarshal error:", err, "body:", req.Body)
		return response(400, map[string]string{"error": "invalid json"})
	}

	// Trim whitespace
	user.UserID = strings.TrimSpace(user.UserID)
	user.Name = strings.TrimSpace(user.Name)
	user.Email = strings.TrimSpace(user.Email)
	user.Password = strings.TrimSpace(user.Password)

	if user.UserID == "" || user.Name == "" || user.Email == "" || user.Password == "" {
		return response(400, map[string]string{"error": "missing fields"})
	}

	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		log.Println("marshal error:", err)
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

	return response(201, map[string]string{"message": "user created"})
}

//////////////////////
// RESPONSE HELPER
//////////////////////

func response(code int, body any) (events.APIGatewayV2HTTPResponse, error) {

	jsonBody, err := json.Marshal(body)
	if err != nil {
		log.Println("json marshal error:", err)

		return events.APIGatewayV2HTTPResponse{
			StatusCode: 500,
			Headers: map[string]string{
				"Content-Type":                "application/json",
				"Access-Control-Allow-Origin": "*",
			},
			Body: `{"error":"json marshal failed"}`,
		}, nil
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode: code,
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "GET,POST,OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type",
		},
		Body: string(jsonBody),
	}, nil
}

//////////////////////
// MAIN
//////////////////////

func main() {
	lambda.Start(handler)
}
