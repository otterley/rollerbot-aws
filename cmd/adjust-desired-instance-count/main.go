package main

import (
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/otterley/rollerbot-aws/internal"
	"github.com/pkg/errors"
)

func adjustDesiredInstanceCount(request internal.RollerParameters) (response internal.RollerParameters, err error) {
	response = request
	sess := session.Must(session.NewSession())
	client := autoscaling.New(sess)

	asgInfo, err := client.DescribeAutoScalingGroups(
		&autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: aws.StringSlice([]string{request.AutoScalingGroupName}),
		},
	)
	if err != nil {
		return response, errors.WithMessage(err, "DescribeAutoScalingGroups")
	}
	if len(asgInfo.AutoScalingGroups) != 1 {
		return response, errors.New("Assertion failure: DescribeAutoScalingGroups did not return exactly 1 group")
	}

	desiredCapacity := aws.Int64Value(asgInfo.AutoScalingGroups[0].DesiredCapacity) +
		int64(request.StepSize)

	fmt.Printf("Adjusting desired capacity on Auto Scaling Group %s from %d to %d\n",
		request.AutoScalingGroupName, asgInfo.AutoScalingGroups[0].DesiredCapacity, desiredCapacity)

	_, err = client.UpdateAutoScalingGroup(
		&autoscaling.UpdateAutoScalingGroupInput{
			AutoScalingGroupName: aws.String(request.AutoScalingGroupName),
			DesiredCapacity:      aws.Int64(desiredCapacity),
		},
	)
	if err != nil {
		return response, errors.WithMessage(err, "UpdateAutoScalingGroup")
	}
	return
}

func main() {
	lambda.Start(adjustDesiredInstanceCount)
}
