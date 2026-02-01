"""
API Lambda handler for YouTube Summary Service.

Provides REST API endpoints for the frontend:
- GET /summaries?hashtag={hashtag} - Get summaries for a hashtag
- GET /hashtags - Get list of available hashtags
"""

import json
import logging
import os
from decimal import Decimal
from typing import Any

import boto3
from boto3.dynamodb.conditions import Key

# Setup logging
logger = logging.getLogger()
logger.setLevel(logging.INFO)

# AWS clients
dynamodb = boto3.resource("dynamodb")


class DecimalEncoder(json.JSONEncoder):
    """Custom JSON encoder for Decimal types from DynamoDB."""

    def default(self, obj):
        if isinstance(obj, Decimal):
            return int(obj) if obj % 1 == 0 else float(obj)
        return super().default(obj)


def get_summaries(table, hashtag: str, limit: int = 50) -> list:
    """
    Get summaries for a specific hashtag, ordered by processedAt descending.
    """
    try:
        summaries = []
        last_key = None
        
        # Define fields to project (exclude large transcript)
        projection = "videoId, title, #sum, detailSummary, processedAt, publishedAt, channelTitle, viewCount, likeCount, thumbnailUrl, thumbnails"
        expr_names = {"#sum": "summary"}

        while True:
            query_params = {
                "KeyConditionExpression": Key("hashtag").eq(hashtag),
                "ScanIndexForward": False,  # Descending order
                "ProjectionExpression": projection,
                "ExpressionAttributeNames": expr_names,
            }
            if last_key:
                query_params["ExclusiveStartKey"] = last_key
            
            if limit > 0:
                remaining = limit - len(summaries)
                if remaining <= 0:
                    break
                query_params["Limit"] = remaining

            response = table.query(**query_params)
            summaries.extend(response.get("Items", []))
            
            last_key = response.get("LastEvaluatedKey")
            if not last_key:
                break
                
        return summaries
    except Exception as e:
        logger.error(f"Error querying summaries: {e}")
        return []


def get_all_hashtags(env_hashtags: list) -> list:
    """Return list of configured hashtags."""
    return env_hashtags


def create_response(status_code: int, body: Any) -> dict:
    """Create API Gateway response with CORS headers."""
    return {
        "statusCode": status_code,
        "headers": {
            "Content-Type": "application/json",
            "Access-Control-Allow-Origin": "*",
            "Access-Control-Allow-Methods": "GET, OPTIONS",
            "Access-Control-Allow-Headers": "Content-Type, Authorization",
        },
        "body": json.dumps(body, cls=DecimalEncoder, ensure_ascii=False),
    }


def lambda_handler(event, context):
    """Main Lambda handler for API Gateway requests."""
    logger.info(f"Received event: {json.dumps(event)}")

    # Get configuration from environment
    table_name = os.environ.get("DYNAMODB_TABLE", "youtube-summary-dev")
    hashtags_json = os.environ.get("HASHTAGS", '["プログラミング", "エンジニア", "Python"]')

    try:
        hashtags = json.loads(hashtags_json)
    except json.JSONDecodeError:
        hashtags = ["プログラミング", "エンジニア", "Python"]

    table = dynamodb.Table(table_name)

    # Parse request
    http_method = event.get("requestContext", {}).get("http", {}).get("method", "GET")
    raw_path = event.get("rawPath", "/")
    query_params = event.get("queryStringParameters") or {}

    # Handle OPTIONS (CORS preflight)
    if http_method == "OPTIONS":
        return create_response(200, {"message": "OK"})

    # Route handling
    if raw_path == "/api/summaries":
        hashtag = query_params.get("hashtag")

        if not hashtag:
            return create_response(400, {"error": "Missing required parameter: hashtag"})

        if hashtag not in hashtags:
            return create_response(400, {"error": f"Invalid hashtag. Available: {hashtags}"})

        limit = int(query_params.get("limit", 0))
        summaries = get_summaries(table, hashtag, limit)

        return create_response(200, {
            "hashtag": hashtag,
            "count": len(summaries),
            "summaries": summaries,
        })

    elif raw_path == "/api/hashtags":
        return create_response(200, {
            "hashtags": hashtags,
        })

    else:
        return create_response(404, {"error": "Not found"})
