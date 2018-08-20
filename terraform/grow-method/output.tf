output "start_roller_lambda_arn" {
  value = "${aws_lambda_function.start_roller.arn}"
}

output "step_function_arn" {
  value = "${aws_sfn_state_machine.roller.id}"
}
