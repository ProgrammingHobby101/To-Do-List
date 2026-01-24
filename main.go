package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	lambda.Start(handler)
}

func handleHello() events.LambdaFunctionURLResponse {
	message := "hello world!"
	return events.LambdaFunctionURLResponse{StatusCode: 200, Body: message}
}

func handleGreet(apiKey string) events.LambdaFunctionURLResponse {
	greeting := "Hi!"
	if apiKey != "valid_key" { // Optional: Validate API Key
		return events.LambdaFunctionURLResponse{StatusCode: 401, Body: "Unauthorized"}
	}
	return events.LambdaFunctionURLResponse{StatusCode: 200, Body: greeting + " you are allowed"}
}

func handler(request events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	path := request.RequestContext.HTTP.Path
	apiKey := request.Headers["x-api-key"]

	var response events.LambdaFunctionURLResponse

	switch path {
	case "/api/health": //HEAD method
		response = handleHello()
	case "/api/to-do-list": //POST method
		response = handleGreet(apiKey)
	default:
		response = events.LambdaFunctionURLResponse{StatusCode: 404, Body: "Not Found"}
	}

	return response, nil
}
