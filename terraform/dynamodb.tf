resource "aws_dynamodb_table" "summaries" {
  name         = "youtube-summary-${local.env}"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "hashtag"
  range_key    = "processedAt"

  attribute {
    name = "hashtag"
    type = "S"
  }

  attribute {
    name = "processedAt"
    type = "S"
  }

  attribute {
    name = "videoId"
    type = "S"
  }

  # GSI for checking duplicate videoId
  global_secondary_index {
    name            = "videoId-index"
    hash_key        = "videoId"
    projection_type = "KEYS_ONLY"
  }

  point_in_time_recovery {
    enabled = local.env == "prd" ? true : false
  }

  tags = {
    Name = "youtube-summary-${local.env}"
  }
}
