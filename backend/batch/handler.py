"""
Batch processing Lambda handler for YouTube video summarization.

This Lambda is triggered every 3 hours by EventBridge to:
1. Search YouTube for videos with specified hashtags
2. Filter by view count and like count
3. Get transcripts using youtube_transcript_api
4. Generate summaries using Gemini 2.5 Flash
5. Store results in DynamoDB
"""

import json
import logging
import os
from datetime import datetime, timezone
from typing import Optional

import boto3
from googleapiclient.discovery import build
from youtube_transcript_api import YouTubeTranscriptApi
from youtube_transcript_api._errors import TranscriptsDisabled, NoTranscriptFound
import google.generativeai as genai

from config import (
    DEFAULT_HASHTAGS,
    MIN_VIEW_COUNT,
    MIN_LIKE_COUNT,
    MAX_RESULTS_PER_HASHTAG,
    SEARCH_ORDER,
    SUMMARY_MAX_CHARS,
)

# Setup logging
logger = logging.getLogger()
logger.setLevel(logging.INFO)

# AWS clients
dynamodb = boto3.resource("dynamodb")
secrets_client = boto3.client("secretsmanager")


def get_secret(secret_name: str) -> str:
    """Retrieve secret value from AWS Secrets Manager."""
    response = secrets_client.get_secret_value(SecretId=secret_name)
    return response["SecretString"]


def get_youtube_client(api_key: str):
    """Create YouTube Data API client."""
    return build("youtube", "v3", developerKey=api_key)


def search_videos_by_hashtag(youtube, hashtag: str, max_results: int = MAX_RESULTS_PER_HASHTAG) -> list:
    """
    Search for videos with a specific hashtag.
    Returns list of video IDs.
    """
    try:
        search_response = youtube.search().list(
            q=f"#{hashtag}",
            part="id",
            type="video",
            order=SEARCH_ORDER,
            maxResults=max_results,
        ).execute()

        return [item["id"]["videoId"] for item in search_response.get("items", [])]
    except Exception as e:
        logger.error(f"Error searching videos for #{hashtag}: {e}")
        return []


def get_video_details(youtube, video_ids: list) -> dict:
    """
    Get video details (title, view count, like count, published date).
    Returns dict keyed by video ID.
    """
    if not video_ids:
        return {}

    try:
        response = youtube.videos().list(
            id=",".join(video_ids),
            part="snippet,statistics",
        ).execute()

        videos = {}
        for item in response.get("items", []):
            video_id = item["id"]
            snippet = item.get("snippet", {})
            statistics = item.get("statistics", {})

            videos[video_id] = {
                "title": snippet.get("title", ""),
                "channelTitle": snippet.get("channelTitle", ""),
                "publishedAt": snippet.get("publishedAt", ""),
                "thumbnails": snippet.get("thumbnails", {}),
                "viewCount": int(statistics.get("viewCount", 0)),
                "likeCount": int(statistics.get("likeCount", 0)),
            }

        return videos
    except Exception as e:
        logger.error(f"Error getting video details: {e}")
        return {}


def get_transcript(video_id: str) -> Optional[str]:
    """
    Get video transcript (auto-generated or manual).
    Returns full transcript text or None if unavailable.
    """
    try:
        transcript_list = YouTubeTranscriptApi.list_transcripts(video_id)

        # Try to get Japanese transcript first, then any available
        try:
            transcript = transcript_list.find_transcript(["ja"])
        except NoTranscriptFound:
            try:
                transcript = transcript_list.find_generated_transcript(["ja"])
            except NoTranscriptFound:
                # Fallback to any available transcript
                transcript = next(iter(transcript_list))

        transcript_data = transcript.fetch()
        full_text = " ".join([entry["text"] for entry in transcript_data])
        return full_text

    except (TranscriptsDisabled, NoTranscriptFound) as e:
        logger.info(f"No transcript available for video {video_id}: {e}")
        return None
    except Exception as e:
        logger.error(f"Error getting transcript for video {video_id}: {e}")
        return None


def generate_summary(transcript: str, title: str, gemini_api_key: str) -> Optional[str]:
    """
    Generate summary using Gemini 2.5 Flash.
    Returns summary text (~400 characters in Japanese).
    """
    try:
        genai.configure(api_key=gemini_api_key)
        model = genai.GenerativeModel("gemini-2.5-flash")

        prompt = f"""以下のYouTube動画の字幕テキストを元に、日本語で{SUMMARY_MAX_CHARS}文字程度の要約を作成してください。

動画タイトル: {title}

字幕テキスト:
{transcript[:10000]}  # Limit transcript length to avoid token limits

要約の要件:
- 動画の主要な内容やポイントを簡潔にまとめる
- 視聴者が動画を見るべきかどうか判断できる情報を含める
- 専門用語があれば適度に説明を加える
- 約{SUMMARY_MAX_CHARS}文字程度に収める

要約:"""

        response = model.generate_content(prompt)
        return response.text.strip()

    except Exception as e:
        logger.error(f"Error generating summary: {e}")
        return None


def is_video_processed(table, video_id: str) -> bool:
    """Check if video has already been processed."""
    try:
        response = table.query(
            IndexName="videoId-index",
            KeyConditionExpression="videoId = :vid",
            ExpressionAttributeValues={":vid": video_id},
            Limit=1,
        )
        return len(response.get("Items", [])) > 0
    except Exception as e:
        logger.error(f"Error checking if video is processed: {e}")
        return False


def save_summary(table, hashtag: str, video_id: str, video_data: dict, summary: str):
    """Save summary to DynamoDB."""
    try:
        now = datetime.now(timezone.utc).isoformat()
        table.put_item(
            Item={
                "hashtag": hashtag,
                "processedAt": now,
                "videoId": video_id,
                "title": video_data["title"],
                "channelTitle": video_data["channelTitle"],
                "publishedAt": video_data["publishedAt"],
                "thumbnails": video_data["thumbnails"],
                "viewCount": video_data["viewCount"],
                "likeCount": video_data["likeCount"],
                "summary": summary,
            }
        )
        logger.info(f"Saved summary for video {video_id}")
    except Exception as e:
        logger.error(f"Error saving summary to DynamoDB: {e}")


def lambda_handler(event, context):
    """Main Lambda handler."""
    logger.info("Starting batch processing")

    # Get configuration from environment
    table_name = os.environ.get("DYNAMODB_TABLE", "youtube-summary-dev")
    hashtags_json = os.environ.get("HASHTAGS", json.dumps(DEFAULT_HASHTAGS))
    min_view_count = int(os.environ.get("MIN_VIEW_COUNT", MIN_VIEW_COUNT))
    min_like_count = int(os.environ.get("MIN_LIKE_COUNT", MIN_LIKE_COUNT))
    youtube_secret_name = os.environ.get("YOUTUBE_API_SECRET", "youtube-summary/youtube-api-key")
    gemini_secret_name = os.environ.get("GEMINI_API_SECRET", "youtube-summary/gemini-api-key")

    # Parse hashtags
    try:
        hashtags = json.loads(hashtags_json)
    except json.JSONDecodeError:
        hashtags = DEFAULT_HASHTAGS

    # Get API keys from Secrets Manager
    youtube_api_key = get_secret(youtube_secret_name)
    gemini_api_key = get_secret(gemini_secret_name)

    # Initialize clients
    youtube = get_youtube_client(youtube_api_key)
    table = dynamodb.Table(table_name)

    # Track statistics
    stats = {
        "videos_found": 0,
        "videos_filtered": 0,
        "videos_without_transcript": 0,
        "videos_already_processed": 0,
        "videos_summarized": 0,
        "errors": 0,
    }

    for hashtag in hashtags:
        logger.info(f"Processing hashtag: #{hashtag}")

        # Search for videos
        video_ids = search_videos_by_hashtag(youtube, hashtag)
        stats["videos_found"] += len(video_ids)

        if not video_ids:
            continue

        # Get video details
        videos = get_video_details(youtube, video_ids)

        for video_id, video_data in videos.items():
            # Filter by view count and like count
            if video_data["viewCount"] < min_view_count:
                logger.info(f"Skipping video {video_id}: view count {video_data['viewCount']} < {min_view_count}")
                stats["videos_filtered"] += 1
                continue

            if video_data["likeCount"] < min_like_count:
                logger.info(f"Skipping video {video_id}: like count {video_data['likeCount']} < {min_like_count}")
                stats["videos_filtered"] += 1
                continue

            # Check if already processed
            if is_video_processed(table, video_id):
                logger.info(f"Skipping video {video_id}: already processed")
                stats["videos_already_processed"] += 1
                continue

            # Get transcript
            transcript = get_transcript(video_id)
            if not transcript:
                stats["videos_without_transcript"] += 1
                continue

            # Generate summary
            summary = generate_summary(transcript, video_data["title"], gemini_api_key)
            if not summary:
                stats["errors"] += 1
                continue

            # Save to DynamoDB
            save_summary(table, hashtag, video_id, video_data, summary)
            stats["videos_summarized"] += 1

    logger.info(f"Batch processing complete. Stats: {json.dumps(stats)}")

    return {
        "statusCode": 200,
        "body": json.dumps({
            "message": "Batch processing complete",
            "stats": stats,
        }),
    }
