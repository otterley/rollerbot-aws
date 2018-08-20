package internal

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/pkg/errors"
)

func CountOutdatedAutoScalingInstances(sess client.ConfigProvider, autoScalingGroupName string) (count int, err error) {
	client := autoscaling.New(sess)

	// Determine Launch Configuration Name
	result, err := client.DescribeAutoScalingGroups(
		&autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: aws.StringSlice([]string{autoScalingGroupName}),
		},
	)
	if err != nil {
		return 0, errors.WithMessage(err, "DescribeAutoScalingGroups")
	}
	if len(result.AutoScalingGroups) == 0 {
		return 0, errors.Errorf("Auto Scaling Group %s not found", autoScalingGroupName)
	}
	group := result.AutoScalingGroups[0]

	// Iterate through instances
	for _, instance := range group.Instances {
		if aws.StringValue(instance.LaunchConfigurationName) != aws.StringValue(group.LaunchConfigurationName) {
			count++
		}
	}
	return
}
