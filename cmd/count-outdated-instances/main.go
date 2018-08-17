package main

import (
	"github.com/aws/aws-lambda-go/lambda"
)

func countOutdatedInstances(request interface{}) (interface{}, error) {
	return nil, nil
}

func main() {
	lambda.Start(countOutdatedInstances)
}
