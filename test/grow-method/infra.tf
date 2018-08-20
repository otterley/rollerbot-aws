provider "aws" {}

variable "lambda_version" {
  type = "string"
}

variable "wait_interval" {
  type    = "string"
  default = "600"
}

module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "1.37.0"

  name = "test-rollerbot-grow-method"
  cidr = "10.0.0.0/16"

  azs            = ["us-west-2a"]
  public_subnets = ["10.0.0.0/24"]

  tags = {
    Test = "rollerbot-grow-method"
  }
}

module "security_group" {
  source  = "terraform-aws-modules/security-group/aws//modules/ssh"
  version = "2.1.0"

  name   = "test-rollerbot-grow-method-ssh"
  vpc_id = "${module.vpc.vpc_id}"

  ingress_cidr_blocks = ["0.0.0.0/0"]
}

data "aws_ami" "amazon_linux" {
  most_recent = true

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }

  filter {
    name   = "owner-alias"
    values = ["amazon"]
  }
}

module "asg" {
  source  = "terraform-aws-modules/autoscaling/aws"
  version = "2.7.0"

  name                        = "test-rollerbot-grow-method"
  image_id                    = "${data.aws_ami.amazon_linux.image_id}"
  instance_type               = "t2.micro"
  health_check_type           = "EC2"
  security_groups             = ["${module.security_group.this_security_group_id}"]
  vpc_zone_identifier         = ["${module.vpc.public_subnets}"]
  associate_public_ip_address = false

  desired_capacity     = 1
  min_size             = 1
  max_size             = 5
  termination_policies = ["OldestInstance"]

  wait_for_capacity_timeout = 0
}

resource "aws_autoscaling_policy" "target_tracking" {
  name                   = "test-rollerbot-grow-method"
  autoscaling_group_name = "${module.asg.this_autoscaling_group_name}"
  policy_type            = "TargetTrackingScaling"

  target_tracking_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ASGAverageCPUUtilization"
    }

    target_value = 90
  }
}

module "roller" {
  source = "../../terraform/grow-method"

  lambda_version = "${var.lambda_version}"

  autoscaling_group_name = "${module.asg.this_autoscaling_group_name}"
  step_size              = 1
  wait_interval          = "${var.wait_interval}"
}

output "launch_configuration_name" {
  value = "${module.asg.this_launch_configuration_name}"
}

output "start_roller_lambda_arn" {
  value = "${module.roller.start_roller_lambda_arn}"
}

output "step_function_arn" {
  value = "${module.roller.step_function_arn}"
}

output "autoscaling_group_name" {
  value = "${module.asg.this_autoscaling_group_name}"
}
