



resource "aws_lambda_function" "api" {
  function_name    = "youtube-summary-api-${local.env}"
  role             = aws_iam_role.api_lambda.arn
  handler          = "bootstrap"
  runtime          = "provided.al2023"
  memory_size      = 256
  timeout          = 30
  filename         = "${path.module}/../.build/api_lambda.zip"
  source_code_hash = filebase64sha256("${path.module}/../.build/api_lambda.zip")

  environment {
    variables = {
      ENVIRONMENT    = local.env
      DYNAMODB_TABLE = aws_dynamodb_table.summaries.name
      CHANNEL_ID     = var.channel_id
    }
  }

  tags = {
    Name = "youtube-summary-api-${local.env}"
  }
}



# CloudWatch Log Groups


resource "aws_cloudwatch_log_group" "api_lambda" {
  name              = "/aws/lambda/${aws_lambda_function.api.function_name}"
  retention_in_days = local.env == "prd" ? 30 : 7
}
