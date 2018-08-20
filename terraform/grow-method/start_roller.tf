resource "aws_lambda_function" "start_roller" {
  function_name = "${format("%.64s", "rollerbot-start-roller-${var.autoscaling_group_name}")}"
  description   = "Start instance roller for ${var.autoscaling_group_name} Auto Scaling Group"
  role          = "${aws_iam_role.start_roller.arn}"

  s3_bucket = "${var.s3_bucket}"
  s3_key    = "v${var.lambda_version}/start-roller.zip"
  handler   = "start-roller"
  runtime   = "go1.x"

  environment {
    variables = {
      AUTOSCALING_GROUP_NAME = "${var.autoscaling_group_name}"
      STATE_MACHINE_ARN      = "${aws_sfn_state_machine.roller.id}"
      STEP_SIZE              = "${var.step_size}"
      STEP_PERCENT           = "${var.step_percent}"
    }
  }
}

data "aws_iam_policy_document" "start_roller_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

data "aws_iam_policy_document" "start_roller_policy" {
  statement {
    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]

    resources = ["*"]
  }

  statement {
    actions   = ["autoscaling:DescribeAutoScalingGroups"]
    resources = ["*"]
  }

  statement {
    actions   = ["states:StartExecution"]
    resources = ["${aws_sfn_state_machine.roller.id}"]
  }
}

resource "aws_iam_role" "start_roller" {
  name               = "${format("%.64s", "rollerbot-start-roller-${var.autoscaling_group_name}")}"
  assume_role_policy = "${data.aws_iam_policy_document.start_roller_assume_role.json}"
}

resource "aws_iam_role_policy" "start_roller" {
  name   = "start-roller"
  role   = "${aws_iam_role.start_roller.name}"
  policy = "${data.aws_iam_policy_document.start_roller_policy.json}"
}
