resource "aws_sfn_state_machine" "roller" {
  name     = "rollerbot-${var.autoscaling_group_name}"
  role_arn = "${aws_iam_role.roller.arn}"

  definition = <<EOF
{
    "Comment": "Rollerbot - ${var.autoscaling_group_name}",
    "StartAt": "CountOutdatedInstances",
    "States": {
        "CountOutdatedInstances": {
            "Type": "Task",
            "Resource": "",
            "End": true
        }
    }
}
EOF
}

data "aws_iam_policy_document" "roller_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "AWS"
      identifiers = ["states.amazonaws.com"]
    }
  }
}

data "aws_iam_policy_document" "roller_policy" {
  statement {
    actions   = []
    resources = []
  }
}

resource "aws_iam_role" "roller" {
  name               = "rollerbot-roller-${var.autoscaling_group_name}"
  assume_role_policy = "${data.aws_iam_policy_document.roller_assume_role.json}"
}

resource "aws_iam_role_policy" "roller" {
  name   = "rollerbot-roller"
  role   = "${aws_iam_role.roller.name}"
  policy = "${data.aws_iam_policy_document.roller_policy.json}"
}
