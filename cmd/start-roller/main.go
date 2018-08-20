package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/otterley/rollerbot-aws/internal"
	"github.com/pkg/errors"
)

func startRoller(input internal.CloudwatchEvent) error {
	stateMachineARN := os.Getenv("STATE_MACHINE_ARN")
	autoScalingGroupName := input.Detail.RequestParameters.AutoScalingGroupName

	requestedStepSize, err := strconv.Atoi(os.Getenv("STEP_SIZE"))
	if err != nil {
		return errors.Errorf("Atoi: Could not convert STEP_SIZE %s to int", os.Getenv("STEP_SIZE"))
	}
	if requestedStepSize < 0 {
		return errors.Errorf("Invalid STEP_SIZE %d, must be >= 0", requestedStepSize)
	}

	requestedStepPercent, err := strconv.ParseFloat(os.Getenv("STEP_PERCENT"), 64)
	if err != nil {
		return errors.Errorf("Atoi: Could not convert STEP_PERCENT %s to float64", os.Getenv("STEP_PERCENT"))
	}
	if requestedStepPercent < 0 || requestedStepPercent > 100 {
		return errors.Errorf("Invalid STEP_PERCENT %d, must be >= 0", requestedStepPercent)
	}

	if input.ErrorCode != "" {
		fmt.Printf("UpdateAutoScalingGroups request returned error: %s - skipping\n", input.ErrorCode)
		return nil
	}

	if input.Detail.RequestParameters.LaunchConfigurationName == "" {
		fmt.Println("No LaunchConfigurationName was specified in UpdateAutoScalingGroups request - skipping")
		return nil
	}

	if autoScalingGroupName != os.Getenv("AUTOSCALING_GROUP_NAME") {
		return errors.Errorf("Assertion failed: AUTOSCALING_GROUP_NAME is %s, but request had %s", os.Getenv("AUTOSCALING_GROUP_NAME"), autoScalingGroupName)
	}

	sess := session.Must(session.NewSession())
	outdatedCount, err := internal.CountOutdatedAutoScalingInstances(sess, autoScalingGroupName)
	if err != nil {
		return errors.WithMessage(err, "CountOutdatedAutoScalingInstances")
	}
	if outdatedCount < 1 {
		fmt.Printf("No outdated instances found for Auto Scaling Group %s - skipping\n", autoScalingGroupName)
		return nil
	}

	startTime := time.Now()
	executionName := startTime.Format("20060102T150405Z0700")

	sfnInput, err := json.Marshal(
		internal.RollerParameters{
			StateMachineARN:       stateMachineARN,
			AutoScalingGroupName:  autoScalingGroupName,
			StartTime:             startTime.Format(time.RFC3339),
			StepSize:              calculateStepSize(outdatedCount, requestedStepSize, requestedStepPercent),
			OutdatedInstanceCount: outdatedCount,
		},
	)
	if err != nil {
		return errors.WithMessage(err, "Error marshaling JSON")
	}

	client := sfn.New(sess)
	_, err = client.StartExecution(&sfn.StartExecutionInput{
		Name:            aws.String(executionName),
		StateMachineArn: aws.String(stateMachineARN),
		Input:           aws.String(string(sfnInput)),
	})
	if err != nil {
		return errors.WithMessage(err, "StartExecution")
	}

	fmt.Printf("Started Step Function %s with execution name %s\n", stateMachineARN, executionName)
	fmt.Printf("Input:\n%s\n", sfnInput)
	return nil
}

func calculateStepSize(outdatedCount, requestedStepSize int, requestedStepPercent float64) int {
	if requestedStepSize == 0 && requestedStepPercent == 0 {
		return 1
	}
	if requestedStepSize > 0 {
		return requestedStepSize
	}
	return int(math.Floor(float64(outdatedCount) * (requestedStepPercent / 100)))
}

func main() {
	lambda.Start(startRoller)
}
