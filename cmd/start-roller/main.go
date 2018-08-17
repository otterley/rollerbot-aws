package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/davecgh/go-spew/spew"
)

func startRoller(input map[string]interface{}) error {
	spew.Dump(input)
	return nil
}

func main() {
	lambda.Start(startRoller)
}
