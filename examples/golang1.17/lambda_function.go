package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
)

type MyEvent struct {
	Name  string `json:"name"`
	Error string `json:"error"`
}

func HandleRequest(ctx context.Context, name MyEvent) (string, error) {
	if name.Error != "" {
		return "", fmt.Errorf("%s", name.Error)
	}
	return fmt.Sprintf("Hello %s!", name.Name), nil
}

func main() {
	fmt.Println("loop start")
	lambda.Start(HandleRequest)
	fmt.Println("loop stop")
}
