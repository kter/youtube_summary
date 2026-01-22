# Reference to existing secrets (created manually)
data "aws_secretsmanager_secret" "youtube_api_key" {
  name = "youtube-summary/youtube-api-key"
}

data "aws_secretsmanager_secret" "gemini_api_key" {
  name = "youtube-summary/gemini-api-key"
}
