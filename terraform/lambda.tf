resource "aws_lambda_function" "batch" {
  function_name    = "youtube-summary-batch-${local.env}"
  role             = aws_iam_role.batch_lambda.arn
  handler          = "bootstrap"
  runtime          = "provided.al2023"
  memory_size      = var.lambda_memory_size
  timeout          = var.lambda_timeout
  filename         = "${path.module}/../.build/batch_lambda.zip"
  source_code_hash = filebase64sha256("${path.module}/../.build/batch_lambda.zip")


  environment {
    variables = {
      ENVIRONMENT        = local.env
      DYNAMODB_TABLE     = aws_dynamodb_table.summaries.name
      CHANNEL_ID         = var.channel_id
      MIN_VIEW_COUNT     = var.min_view_count
      MIN_LIKE_COUNT     = var.min_like_count
      YOUTUBE_API_SECRET = data.aws_secretsmanager_secret.youtube_api_key.name
    }
  }

  tags = {
    Name = "youtube-summary-batch-${local.env}"
  }
}



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
resource "aws_cloudwatch_log_group" "batch_lambda" {
  name              = "/aws/lambda/${aws_lambda_function.batch.function_name}"
  retention_in_days = local.env == "prd" ? 30 : 7
}

resource "aws_cloudwatch_log_group" "api_lambda" {
  name              = "/aws/lambda/${aws_lambda_function.api.function_name}"
  retention_in_days = local.env == "prd" ? 30 : 7
}
