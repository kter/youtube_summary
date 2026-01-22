variable "channel_id" {
  description = "YouTube channel ID to monitor (@noiehoie)"
  type        = string
  default     = "UC2kM01yXNnouBsJJ0ghyfMg"
}

variable "min_view_count" {
  description = "Minimum view count for filtering"
  type        = number
  default     = 0
}

variable "min_like_count" {
  description = "Minimum like count for filtering"
  type        = number
  default     = 0
}

variable "batch_schedule" {
  description = "Cron expression for batch processing (every 3 hours)"
  type        = string
  default     = "rate(3 hours)"
}

variable "lambda_memory_size" {
  description = "Memory size for Lambda functions"
  type        = number
  default     = 512
}

variable "lambda_timeout" {
  description = "Timeout for Lambda functions in seconds"
  type        = number
  default     = 300
}


