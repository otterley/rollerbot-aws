resource "aws_lambda_function" "adjust_desired_instance_count" {
  function_name = "${format("%.64s", "rollerbot-adjust_count-${var.autoscaling_group_name}")}"
  description   = "Rollerbot - adjust-desired-instance-count for ${var.autoscaling_group_name} Auto Scaling Group"
  role          = "${aws_iam_role.adjust_desired_instance_count.arn}"

  s3_bucket = "${var.s3_bucket}"
  s3_key    = "v${var.lambda_version}/adjust-desired-instance-count.zip"
  handler   = "adjust-desired-instance-count"
  runtime   = "go1.x"
}

data "aws_iam_policy_document" "adjust_desired_instance_count_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

data "aws_iam_policy_document" "adjust_desired_instance_count_policy" {
  statement {
    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]

    resources = ["*"]
  }

  statement {
    actions = [
      "autoscaling:DescribeAutoScalingGroups",
      "autoscaling:UpdateAutoScalingGroup",
    ]

    resources = ["*"]
  }
}

resource "aws_iam_role" "adjust_desired_instance_count" {
  name               = "${format("%.64s", "rollerbot-adjust_count-${var.autoscaling_group_name}")}"
  assume_role_policy = "${data.aws_iam_policy_document.adjust_desired_instance_count_assume_role.json}"
}

resource "aws_iam_role_policy" "adjust_desired_instance_count" {
  name   = "adjust_desired_instance_count"
  role   = "${aws_iam_role.adjust_desired_instance_count.name}"
  policy = "${data.aws_iam_policy_document.adjust_desired_instance_count_policy.json}"
}
