"""
Configuration for YouTube Summary batch processing.
"""

# Hashtags to monitor (can be overridden by environment variable)
DEFAULT_HASHTAGS = [
    "プログラミング",
    "エンジニア",
    "Python",
]

# Filtering thresholds
MIN_VIEW_COUNT = 1000
MIN_LIKE_COUNT = 50

# YouTube API settings
MAX_RESULTS_PER_HASHTAG = 50
SEARCH_ORDER = "date"  # Order by upload date

# Summary settings
SUMMARY_MAX_CHARS = 400
SUMMARY_LANGUAGE = "ja"
