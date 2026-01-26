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
)

type User struct {
	UserID   string `json:"userId" dynamodbav:"userId"`
	Name     string `json:"name" dynamodbav:"name"`
	Email    string `json:"email" dynamodbav:"email"`
	Password string `json:"password" dynamodbav:"password"`
}

var dbClient *dynamodb.Client // DynamoDB client
var tableName1 string = "Users"
var tableName2 string = "Projects"
var tableName3 string = "Project_Items"

// runs on cold-starts. init function to initialize DynamoDB client
// Runs once per container (cold start)
func init() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatal("unable to load AWS SDK config:", err)
	}

	dbClient = dynamodb.NewFromConfig(cfg)
}

func main() {
	lambda.Start(handler)
}

func jsonResponse(code int, body any) (events.APIGatewayV2HTTPResponse, error) {

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 500,
			Body:       `{"error":"json marshal failed"}`,
		}, nil
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode: code,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(jsonBody),
	}, nil
}
func handleHello() events.APIGatewayV2HTTPResponse {
	return events.APIGatewayV2HTTPResponse{
		StatusCode: 204,
	}
}
func createUser(
	ctx context.Context,
	request events.APIGatewayV2HTTPRequest,
) (events.APIGatewayV2HTTPResponse, error) {

	var user User

	// Parse JSON
	if err := json.Unmarshal([]byte(request.Body), &user); err != nil {
		return jsonResponse(400, map[string]string{
			"error": "invalid json",
		})
	}

	// Validate fields
	if user.UserID == "" || user.Name == "" || user.Email == "" {
		return jsonResponse(400, map[string]string{
			"error": "missing fields",
		})
	}

	// Convert to DynamoDB item
	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		log.Println("Marshal error:", err)
		return jsonResponse(500, map[string]string{
			"error": "marshal failed",
		})
	}

	// Write to DynamoDB
	_, err = dbClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName1),
		Item:      item,
	})

	if err != nil {
		log.Println("PutItem error:", err)
		return jsonResponse(500, map[string]string{
			"error": "dynamodb error",
		})
	}

	// Optional API Key check
	apiKey := request.Headers["x-api-key"]
	if apiKey != "valid_key" {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 401,
			Body:       "Unauthorized",
		}, nil
	}

	return jsonResponse(201, map[string]string{
		"message": "User created successfully",
	})
}

// Helper
// func clientError(code int, msg string) (events.LambdaFunctionURLResponse, error) {
// 	body, _ := json.Marshal(map[string]string{"error": msg})
// 	return events.LambdaFunctionURLResponse{
// 		StatusCode: code,
// 		Body:       string(body),
// 	}, nil
// }

// Helper
// func serverError() (events.LambdaFunctionURLResponse, error) {
// 	body, _ := json.Marshal(map[string]string{"error": "Internal Server Error"})
// 	return events.LambdaFunctionURLResponse{
// 		StatusCode: 500,
// 		Body:       string(body),
// 	}, nil
// }

// Helper
// func response(code int, body any) (events.LambdaFunctionURLResponse, error) {
// 	jsonBody, _ := json.Marshal(body)

//		return events.LambdaFunctionURLResponse{
//			StatusCode: code,
//			Headers: map[string]string{
//				"Content-Type": "application/json",
//			},
//			Body: string(jsonBody),
//		}, nil
//	}
func handler(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	path := request.RequestContext.HTTP.Path
	method := request.RequestContext.HTTP.Method

	var response events.APIGatewayV2HTTPResponse

	switch method + " " + path {
	case "HEAD /api/to-do-list/mypost/health": //HEAD method
		return handleHello(), nil
	case "POST /api/to-do-list/mypost/users": //POST method
		response, _ = createUser(ctx, request)
	default:
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 404,
			Body:       "Not Found",
		}, nil
	}

	return response, nil
}
