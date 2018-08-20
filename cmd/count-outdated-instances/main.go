package main

import (
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/otterley/rollerbot-aws/internal"
	"github.com/pkg/errors"
)

func countOutdatedInstances(request internal.RollerParameters) (response internal.RollerParameters, err error) {
	response = request
	sess := session.Must(session.NewSession())

	response.OutdatedInstanceCount, err = internal.CountOutdatedAutoScalingInstances(sess, request.AutoScalingGroupName)
	if err != nil {
		return response, errors.WithMessage(err, "CountOutdatedAutoScalingInstances")
	}

	fmt.Printf("Auto Scaling Group %s has %d outdated instances\n", response.AutoScalingGroupName, response.OutdatedInstanceCount)

	return
}

func main() {
	lambda.Start(countOutdatedInstances)
}
