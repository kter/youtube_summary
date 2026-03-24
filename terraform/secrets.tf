# Parameter Store naming contract for external API keys.
# Values are written outside Terraform to avoid persisting secrets in state.
data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

locals {
  youtube_api_parameter_name = "/youtube-summary/youtube-api-key"
  gemini_api_parameter_name  = "/youtube-summary/gemini-api-key"

  youtube_api_parameter_arn = "arn:aws:ssm:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:parameter${local.youtube_api_parameter_name}"
  gemini_api_parameter_arn  = "arn:aws:ssm:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:parameter${local.gemini_api_parameter_name}"
}
