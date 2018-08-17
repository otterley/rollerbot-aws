resource "aws_cloudwatch_event_rule" "update_group" {
  name        = "rollerbot-update_group-${var.autoscaling_group_name}"
  description = "Invoked when UpdateAutoScalingGroup is called on ${var.autoscaling_group_name}"

  event_pattern = <<PATTERN
{
    "detail-type": [ "AWS API Call via CloudTrail" ],
    "detail": {
        "eventSource": [ "autoscaling.amazonaws.com" ],
        "eventName": [ "UpdateAutoScalingGroup "]
    }
}
PATTERN
}

resource "aws_cloudwatch_event_target" "update_group" {
  rule = "${aws_cloudwatch_event_rule.update_group.name}"
  arn  = "${aws_lambda_function.start_roller.arn}"
}

resource "aws_lambda_permission" "start_roller" {
  statement_id  = "AllowExecutionFromCloudWatch"
  action        = "lambda:InvokeFunction"
  function_name = "${aws_lambda_function.start_roller.function_name}"
  principal     = "events.amazonaws.com"
  source_arn    = "${aws_cloudwatch_event_rule.update_group.arn}"
}
