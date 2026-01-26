package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var (
	dbClient  *dynamodb.Client
	tableName = "To-Do-List-Users"
)

// ✅ Your struct
type User struct {
	UserID   string `json:"userId" dynamodbav:"userId"`
	Name     string `json:"name" dynamodbav:"name"`
	Email    string `json:"email" dynamodbav:"email"`
	Password string `json:"password" dynamodbav:"password"`
}

// Runs once per container (cold start)
func init() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatal("unable to load AWS SDK config:", err)
	}

	dbClient = dynamodb.NewFromConfig(cfg)
}

func handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	//debug
	method_and_path_debug := "request method " + req.RequestContext.HTTP.Method + ". request path;" + req.RequestContext.HTTP.Path
	log.Println("debug for handler: ", method_and_path_debug)

	// prepare for switch
	method := req.RequestContext.HTTP.Method
	path := req.RequestContext.HTTP.Path

	switch method + " " + path {

	case "POST /api/to-do-list/mypost/users": //POST method
		return createUser(ctx, req)
	case "HEAD /api/to-do-list/mypost/health": //HEAD method
		return handleHello(ctx, req)
	case "GET /api/to-do-list/mypost/users": //HEAD method
		return getUser(ctx, req)

	case "OPTIONS /api/to-do-list/mypost/users":
		return response(200, map[string]string{"message": "ok"})

	default:
		return response(405, map[string]string{
			"error": "method not allowed",
		})
	}
}

func handleHello(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return response(204, nil)
}

// ---------------- CREATE USER ----------------

func createUser(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {

	var user User

	if err := json.Unmarshal([]byte(req.Body), &user); err != nil {
		return response(400, map[string]string{"error": "invalid json"})
	}

	if user.UserID == "" || user.Name == "" || user.Email == "" || user.Password == "" {
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

	return response(201, map[string]string{"message": "user created"})
}

// ---------------- GET USER ----------------

func getUser(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {

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

// ---------------- RESPONSE HELPER ----------------

func response(code int, body any) (events.APIGatewayV2HTTPResponse, error) {

	jsonBody, _ := json.Marshal(body)

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

func main() {
	lambda.Start(handler)
}
