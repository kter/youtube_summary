terraform {
  required_version = ">= 1.0.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  backend "s3" {
    bucket         = "terraform-state-youtube-summary"
    key            = "terraform.tfstate"
    region         = "ap-northeast-1"
    dynamodb_table = "terraform-locks"
    encrypt        = true
  }
}

locals {
  env = terraform.workspace

  domain_map = {
    dev = "youtube-summarize.dev.devtools.site"
    prd = "youtube-summarize.devtools.site"
  }

  zone_map = {
    dev = "dev.devtools.site"
    prd = "devtools.site"
  }

  domain = local.domain_map[local.env]
  zone_name = local.zone_map[local.env]

  common_tags = {
    Project     = "youtube-summary"
    Environment = local.env
    ManagedBy   = "terraform"
  }
}

data "aws_route53_zone" "main" {
  name = local.zone_name
}

provider "aws" {
  region  = "ap-northeast-1"
  profile = local.env

  default_tags {
    tags = local.common_tags
  }
}

# For ACM certificates (must be in us-east-1 for CloudFront)
provider "aws" {
  alias   = "us_east_1"
  region  = "us-east-1"
  profile = local.env

  default_tags {
    tags = local.common_tags
  }
}
