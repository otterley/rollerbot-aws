resource "aws_lambda_function" "start_roller" {
  function_name = "rollerbot-start-roller-${var.autoscaling_group_name}"
  description   = "Start instance roller for ${var.autoscaling_group_name} Auto Scaling Group"
  role          = "${aws_iam_role.start_roller.name}"

  s3_bucket = "${var.s3_bucket}"
  s3_key    = "${var.lambda_version}/start-roller.zip"
  handler   = "start-roller"

  environment {
    variables = {
      AUTOSCALING_GROUP_NAME = "${var.autoscaling_group_name}"
      STEP_FUNCTION_ARN      = "${aws_sfn_state_machine.roller.id}"
      STEP_SIZE              = "${var.step_size}"
      STEP_PERCENT           = "${var.step_percent}"
    }
  }
}

data "aws_iam_policy_document" "start_roller_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "AWS"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

data "aws_iam_policy_document" "start_roller_policy" {
  statement {
    actions   = ["sfn:StartExecution"]
    resources = ["${aws_sfn_state_machine.roller.arn}"]
  }
}

resource "aws_iam_role" "start_roller" {
  name               = "rollerbot-start-roller-${var.autoscaling_group_name}"
  assume_role_policy = "${data.aws_iam_policy_document.start_roller_assume_role.json}"
}

resource "aws_iam_role_policy" "start_roller" {
  name   = "start-roller"
  role   = "${aws_iam_role.start_roller.name}"
  policy = "${data.aws_iam_policy_document.start_roller_policy.json}"
}
