import boto3
from googleapiclient.discovery import build

def get_parameter(parameter_name):
    client = boto3.client("ssm", region_name="ap-northeast-1")
    response = client.get_parameter(Name=parameter_name, WithDecryption=True)
    return response["Parameter"]["Value"]

def debug_search():
    try:
        api_key = get_parameter("/youtube-summary/youtube-api-key")
        youtube = build("youtube", "v3", developerKey=api_key)
        
        hashtag = "Python"
        print(f"Searching for #{hashtag}...")
        
        response = youtube.search().list(
            q=f"#{hashtag}",
            part="id,snippet",
            type="video",
            order="date",
            maxResults=10
        ).execute()
        
        print(f"Response items count: {len(response.get('items', []))}")
        for item in response.get("items", []):
            print(f"ID: {item['id']['videoId']}, Title: {item['snippet']['title']}, Published: {item['snippet']['publishedAt']}")
            
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    debug_search()
