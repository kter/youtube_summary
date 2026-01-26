# IAM Role for Batch Lambda


# IAM Role for API Lambda
resource "aws_iam_role" "api_lambda" {
  name = "youtube-summary-api-lambda-${local.env}"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy" "api_lambda" {
  name = "youtube-summary-api-lambda-policy-${local.env}"
  role = aws_iam_role.api_lambda.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Resource = "arn:aws:logs:*:*:*"
      },
      {
        Effect = "Allow"
        Action = [
          "dynamodb:Query",
          "dynamodb:Scan"
        ]
        Resource = [
          aws_dynamodb_table.summaries.arn,
          "${aws_dynamodb_table.summaries.arn}/index/*"
        ]
      }
    ]
  })
}
