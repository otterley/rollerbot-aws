variable "autoscaling_group_name" {
  description = "Name of Auto Scaling Group to be managed"
}

variable "step_size" {
  description = "Number of instances to add to the Auto Scaling Group at each step"
  default     = "0"
}

variable "step_percent" {
  description = "Number of instances, as a percentage of outdated instances (between 1-100), to add to the Auto Scaling Group at each step"
  default     = "0"
}

variable "wait_interval" {
  description = "Number of seconds to wait for the Auto Scaler to scale in the group before scaling out again. 1800 seconds is recommended if a Target Tracking policy is in place."
  default     = "1800"
}

variable "lambda_version" {
  description = "Lambda function version"
}

variable "s3_bucket" {
  description = "S3 bucket in which Lambda functions live"
  default     = "rollerbot-aws"
}
