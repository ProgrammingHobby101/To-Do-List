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
type LoginUser struct {
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
	case "GET /api/to-do-list/mypost/users": //GET method
		return loginUser(ctx, req)
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

	// Trim whitespace from all string fields
	user.UserID = strings.TrimSpace(user.UserID)
	user.Name = strings.TrimSpace(user.Name)
	user.Email = strings.TrimSpace(user.Email)
	user.Password = strings.TrimSpace(user.Password)

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

	return response(201, map[string]string{"message": "user " + user.Name + " created"})
}

// ---------------- GET USER ----------------

func loginUser(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	var login LoginUser

	// Parse JSON body into LoginUser struct
	if err := json.Unmarshal([]byte(req.Body), &login); err != nil {
		return response(400, map[string]string{"error": "invalid JSON"})
	}

	email := login.Email
	password := login.Password
	// protect against accidental whitespace
	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)

	if email == "" || password == "" {
		return response(400, map[string]string{"error": "email and password required"})
	}

	// Scan table for matching email & password
	input := &dynamodb.ScanInput{
		TableName:        aws.String(tableName),
		FilterExpression: aws.String("email = :email AND password = :password"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":email":    &types.AttributeValueMemberS{Value: email},
			":password": &types.AttributeValueMemberS{Value: password},
		},
		Limit: aws.Int32(1),
	}

	result, err := dbClient.Scan(ctx, input)
	if err != nil {
		log.Println("Scan error:", err)
		return response(500, map[string]string{"error": "dynamodb error"})
	}

	if len(result.Items) == 0 {
		// Login failed
		return response(401, map[string]string{"error": "invalid email or password"})
	}

	var user User
	if err := attributevalue.UnmarshalMap(result.Items[0], &user); err != nil {
		log.Println("Unmarshal error:", err)
		return response(500, map[string]string{"error": "unmarshal error"})
	}

	// Remove password before returning
	user.Password = ""

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
