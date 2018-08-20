package internal

type CloudwatchEvent struct {
	Detail    CloudwatchEventDetail `json:"detail"`
	ErrorCode string                `json:"errorCode"`
}

type CloudwatchEventDetail struct {
	RequestParameters UpdateAutoScalingGroupParameters `json:"requestParameters"`
}

type UpdateAutoScalingGroupParameters struct {
	AutoScalingGroupName    string `json:"autoScalingGroupName"`
	LaunchConfigurationName string `json:"launchConfigurationName"`
}

type RollerParameters struct {
	StateMachineARN       string
	AutoScalingGroupName  string
	StartTime             string // RFC3339 format
	StepSize              int
	OutdatedInstanceCount int
	RunningExecutionCount int
}
