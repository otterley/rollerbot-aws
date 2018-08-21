package grow_method_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/otterley/rollerbot-aws/internal"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestGrowMethod(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(20*time.Minute))
	defer cancel()

	tfOpts := &terraform.Options{
		Vars: map[string]interface{}{
			"lambda_version": internal.MustEnv("LAMBDA_VERSION"),
		},
	}
	defer terraform.Destroy(t, tfOpts)
	terraform.InitAndApply(t, tfOpts)

	copiedLCName, err := copyAndAssignLaunchConfig(
		terraform.Output(t, tfOpts, "autoscaling_group_name"),
		terraform.Output(t, tfOpts, "launch_configuration_name"),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer deleteLaunchConfig(copiedLCName)

	t.Run("Cloudwatch Event Connected",
		testCloudwatchEventConnected(ctx, terraform.Output(t, tfOpts, "start_roller_lambda_arn")))
	t.Run("Step Function Started",
		testStepFunctionStarted(ctx, terraform.Output(t, tfOpts, "step_function_arn")))
	t.Run("Auto Scaling Group Grows",
		testAutoScalingGroupGrows(ctx, terraform.Output(t, tfOpts, "autoscaling_group_name")))
	t.Run("Instances have new Launch Configuration",
		testAllInstancesHaveLaunchConfig(ctx, terraform.Output(t, tfOpts, "autoscaling_group_name"), copiedLCName))
	t.Run("Step Function Completes Successfully",
		testStepFunctionOK(ctx, terraform.Output(t, tfOpts, "step_function_arn")))
}

func testCloudwatchEventConnected(ctx context.Context, targetARN string) func(t *testing.T) {
	return func(t *testing.T) {
		client := cloudwatchevents.New(session.Must(session.NewSession()))
		output, err := client.ListRuleNamesByTargetWithContext(
			ctx,
			&cloudwatchevents.ListRuleNamesByTargetInput{
				TargetArn: aws.String(targetARN),
			},
		)
		assert.NoError(t, err)
		assert.NotEmpty(t, output.RuleNames)
	}
}

func testStepFunctionStarted(ctx context.Context, stateMachineARN string) func(t *testing.T) {
	return func(t *testing.T) {
		client := sfn.New(session.Must(session.NewSession()))

		for {
			executions, err := client.ListExecutionsWithContext(
				ctx,
				&sfn.ListExecutionsInput{
					StateMachineArn: aws.String(stateMachineARN),
				},
			)
			assert.NoError(t, err)

			for _, execution := range executions.Executions {
				if aws.StringValue(execution.Status) == "RUNNING" {
					return
				}
				assert.NotContains(t, aws.StringValue(execution.Status), []string{"FAILED", "TIMED_OUT", "ABORTED"})
			}

			fmt.Printf("Waiting for Step Function %s to start\n", stateMachineARN)
			select {
			case <-ctx.Done():
				// timed out
				return
			case <-time.After(10 * time.Second):
				// check again
			}
		}
	}
}

func testAutoScalingGroupGrows(ctx context.Context, autoScalingGroupName string) func(t *testing.T) {
	return func(t *testing.T) {
		client := autoscaling.New(session.Must(session.NewSession()))

		for {
			result, err := client.DescribeAutoScalingGroupsWithContext(
				ctx,
				&autoscaling.DescribeAutoScalingGroupsInput{
					AutoScalingGroupNames: aws.StringSlice([]string{autoScalingGroupName}),
				},
			)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(result.AutoScalingGroups))
			inServiceCount := 0
			for _, instance := range result.AutoScalingGroups[0].Instances {
				if aws.StringValue(instance.LifecycleState) == "InService" {
					inServiceCount++
				}
			}
			fmt.Printf("%d InService instances running in Auto Scaling Group %s\n", inServiceCount, autoScalingGroupName)
			if inServiceCount > 1 {
				return
			}
			select {
			case <-ctx.Done():
				// timed out
				return
			case <-time.After(30 * time.Second):
				// check again
			}
		}
	}
}

func testAllInstancesHaveLaunchConfig(ctx context.Context, autoScalingGroupName, launchConfigurationName string) func(t *testing.T) {
	return func(t *testing.T) {
		client := autoscaling.New(session.Must(session.NewSession()))
		for {
			result, err := client.DescribeAutoScalingGroupsWithContext(
				ctx,
				&autoscaling.DescribeAutoScalingGroupsInput{
					AutoScalingGroupNames: aws.StringSlice([]string{autoScalingGroupName}),
				},
			)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(result.AutoScalingGroups))
			nonMatching := 0
			for _, instance := range result.AutoScalingGroups[0].Instances {
				if aws.StringValue(instance.LaunchConfigurationName) != launchConfigurationName {
					nonMatching++
				}
			}
			if nonMatching == 0 {
				return
			}
			fmt.Printf("%d instances still running with old Launch Configuration\n", nonMatching)
			select {
			case <-ctx.Done():
				// timed out
				return
			case <-time.After(30 * time.Second):
				// check again
			}
		}
	}
}

func testStepFunctionOK(ctx context.Context, stateMachineARN string) func(t *testing.T) {
	return func(t *testing.T) {
		client := sfn.New(session.Must(session.NewSession()))

		for {
			executions, err := client.ListExecutionsWithContext(
				ctx,
				&sfn.ListExecutionsInput{
					StateMachineArn: aws.String(stateMachineARN),
				},
			)
			assert.NoError(t, err)
			for _, execution := range executions.Executions {
				if aws.StringValue(execution.Status) == "SUCCEEDED" {
					return
				}
				assert.NotContains(t, aws.StringValue(execution.Status), []string{"FAILED", "TIMED_OUT", "ABORTED"})
			}
			select {
			case <-ctx.Done():
				// timed out
				return
			case <-time.After(10 * time.Second):
				// check again
			}
		}
	}
}

func copyAndAssignLaunchConfig(autoScalingGroupName, launchConfigurationName string) (string, error) {
	client := autoscaling.New(session.Must(session.NewSession()))
	launchConfigs, err := client.DescribeLaunchConfigurations(
		&autoscaling.DescribeLaunchConfigurationsInput{
			LaunchConfigurationNames: aws.StringSlice([]string{launchConfigurationName}),
		},
	)
	if err != nil {
		return "", err
	}
	if len(launchConfigs.LaunchConfigurations) != 1 {
		return "", fmt.Errorf("Did not find exactly 1 Launch Configuration named %s", launchConfigurationName)
	}
	var input autoscaling.CreateLaunchConfigurationInput
	awsutil.Copy(&input, launchConfigs.LaunchConfigurations[0])
	input.LaunchConfigurationName = aws.String(aws.StringValue(input.LaunchConfigurationName) + "TestCopy")
	// awsutil.Copy sets these to empty strings; the API does not approve of this.
	input.KernelId = nil
	input.KeyName = nil
	input.RamdiskId = nil
	_, err = client.CreateLaunchConfiguration(&input)
	if err != nil {
		return "", err
	}

	_, err = client.UpdateAutoScalingGroup(
		&autoscaling.UpdateAutoScalingGroupInput{
			AutoScalingGroupName:    aws.String(autoScalingGroupName),
			LaunchConfigurationName: input.LaunchConfigurationName,
		},
	)
	return aws.StringValue(input.LaunchConfigurationName), err
}

func deleteLaunchConfig(name string) {
	client := autoscaling.New(session.Must(session.NewSession()))
	client.DeleteLaunchConfiguration(
		&autoscaling.DeleteLaunchConfigurationInput{
			LaunchConfigurationName: aws.String(name),
		},
	)
}
