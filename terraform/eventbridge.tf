# EventBridge Rule for scheduled batch processing
resource "aws_cloudwatch_event_rule" "batch_schedule" {
  name                = "youtube-summary-batch-schedule-${local.env}"
  description         = "Trigger batch processing every 3 hours"
  schedule_expression = var.batch_schedule

  tags = {
    Name = "youtube-summary-batch-schedule-${local.env}"
  }
}

resource "aws_cloudwatch_event_target" "batch_lambda" {
  rule      = aws_cloudwatch_event_rule.batch_schedule.name
  target_id = "youtube-summary-batch-${local.env}"
  arn       = aws_lambda_function.batch.arn
}

resource "aws_lambda_permission" "eventbridge" {
  statement_id  = "AllowEventBridgeInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.batch.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.batch_schedule.arn
}
