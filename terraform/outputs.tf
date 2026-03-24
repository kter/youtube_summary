output "cloudfront_distribution_id" {
  description = "CloudFront distribution ID"
  value       = aws_cloudfront_distribution.frontend.id
}

output "cloudfront_domain_name" {
  description = "CloudFront domain name"
  value       = aws_cloudfront_distribution.frontend.domain_name
}

output "website_url" {
  description = "Website URL"
  value       = "https://${local.domain}"
}

output "api_endpoint" {
  description = "API Gateway endpoint URL"
  value       = aws_apigatewayv2_stage.api.invoke_url
}

output "s3_bucket_name" {
  description = "S3 bucket name for frontend"
  value       = aws_s3_bucket.frontend.id
}



output "api_lambda_function_name" {
  description = "API Lambda function name"
  value       = aws_lambda_function.api.function_name
}

output "dynamodb_table_name" {
  description = "DynamoDB table name"
  value       = aws_dynamodb_table.summaries.name
}

output "youtube_api_parameter_name" {
  description = "SSM Parameter Store name for the YouTube API key"
  value       = local.youtube_api_parameter_name
}

output "gemini_api_parameter_name" {
  description = "SSM Parameter Store name for the Gemini API key"
  value       = local.gemini_api_parameter_name
}
