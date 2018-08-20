resource "aws_sfn_state_machine" "roller" {
  name     = "rollerbot-${var.autoscaling_group_name}"
  role_arn = "${aws_iam_role.roller.arn}"

  definition = <<EOF
{
    "Comment": "Rollerbot - ${var.autoscaling_group_name}",
    "StartAt": "CountRunningExecutions",
    "States": {
        "CountRunningExecutions": {
            "Type": "Task",
            "Resource": "${aws_lambda_function.count_running_executions.arn}",
            "Next": "HaltIfRunningExecutions"
        },
        "HaltIfRunningExecutions": {
            "Type": "Choice",
            "Choices": [
                {
                    "Variable": "$.RunningExecutionCount",
                    "NumericGreaterThan": 1,
                    "Next": "Done"
                }
            ],
            "Default": "CountOutdatedInstances"
        },
        "CountOutdatedInstances": {
            "Type": "Task",
            "Resource": "${aws_lambda_function.count_outdated_instances.arn}",
            "Next": "HaltIfNoneOutdated" 
        },
        "HaltIfNoneOutdated": {
            "Type": "Choice",
            "Choices": [
                {
                    "Variable": "$.OutdatedInstanceCount",
                     "NumericEquals": 0,
                    "Next": "Done"
                }
            ],
            "Default": "IncreaseDesiredCapacity"
        },
        "IncreaseDesiredCapacity": {
            "Type": "Task",
            "Resource": "${aws_lambda_function.adjust_desired_instance_count.arn}",
            "Next": "WaitAndCountAgain"
        },
        "WaitAndCountAgain": {
            "Type": "Wait",
            "Seconds": ${var.wait_interval},
            "Next": "CountOutdatedInstances"
        },
        "Done": {
            "Type": "Succeed"
        }
    }
}
EOF
}

data "aws_iam_policy_document" "roller_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["states.amazonaws.com"]
    }
  }
}

data "aws_iam_policy_document" "roller_policy" {
  statement {
    actions = ["lambda:InvokeFunction"]

    resources = [
      "${aws_lambda_function.count_running_executions.arn}",
      "${aws_lambda_function.count_outdated_instances.arn}",
      "${aws_lambda_function.adjust_desired_instance_count.arn}",
    ]
  }
}

resource "aws_iam_role" "roller" {
  name               = "${format("%.64s", "rollerbot-roller-${var.autoscaling_group_name}")}"
  assume_role_policy = "${data.aws_iam_policy_document.roller_assume_role.json}"
}

resource "aws_iam_role_policy" "roller" {
  name   = "rollerbot-roller"
  role   = "${aws_iam_role.roller.name}"
  policy = "${data.aws_iam_policy_document.roller_policy.json}"
}
