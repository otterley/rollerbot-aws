resource "aws_lambda_function" "count_outdated_instances" {
  function_name = "${format("%.64s", "rollerbot-count_outdated-${var.autoscaling_group_name}")}"
  description   = "Rollerbot - count-outdated-instances for ${var.autoscaling_group_name} Auto Scaling Group"
  role          = "${aws_iam_role.count_outdated_instances.arn}"

  s3_bucket = "${var.s3_bucket}"
  s3_key    = "v${var.lambda_version}/count-outdated-instances.zip"
  handler   = "count-outdated-instances"
  runtime   = "go1.x"
}

data "aws_iam_policy_document" "count_outdated_instances_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

data "aws_iam_policy_document" "count_outdated_instances_policy" {
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
}

resource "aws_iam_role" "count_outdated_instances" {
  name               = "${format("%.64s", "rollerbot-count_outdated-${var.autoscaling_group_name}")}"
  assume_role_policy = "${data.aws_iam_policy_document.count_outdated_instances_assume_role.json}"
}

resource "aws_iam_role_policy" "count_outdated_instances" {
  name   = "count_outdated_instances"
  role   = "${aws_iam_role.count_outdated_instances.name}"
  policy = "${data.aws_iam_policy_document.count_outdated_instances_policy.json}"
}
