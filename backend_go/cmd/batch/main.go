package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/horiagug/youtube-transcript-api-go/pkg/yt_transcript"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

// Global Configuration
var (
	dynamoClient   *dynamodb.Client
	secretsClient  *secretsmanager.Client
	bedrockClient  *bedrockruntime.Client
	tableName      string
	minViewCount   uint64
	minLikeCount   uint64
)

type BatchStats struct {
	VideosFound            int `json:"videos_found"`
	VideosFiltered         int `json:"videos_filtered"`
	VideosWithoutTx        int `json:"videos_without_transcript"`
	VideosAlreadyProcessed int `json:"videos_already_processed"`
	VideosSummarized       int `json:"videos_summarized"`
	Errors                 int `json:"errors"`
}

type VideoDetails struct {
	ID           string
	Title        string
	ChannelTitle string
	PublishedAt  string
	Thumbnails   *youtube.ThumbnailDetails
	ViewCount    uint64
	LikeCount    uint64
}

type SummaryData struct {
	ShortSummary  string `json:"short_summary"`
	DetailSummary string `json:"detail_summary"`
}

func init() {
	// Initialize AWS clients
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-northeast-1"))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	dynamoClient = dynamodb.NewFromConfig(cfg)
	secretsClient = secretsmanager.NewFromConfig(cfg)

	// Bedrock client needs us-east-1 region for Claude models
	bedrockCfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("unable to load Bedrock SDK config, %v", err)
	}
	bedrockClient = bedrockruntime.NewFromConfig(bedrockCfg)
}

func getSecret(ctx context.Context, secretName string) (string, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}
	result, err := secretsClient.GetSecretValue(ctx, input)
	if err != nil {
		return "", err
	}
	if result.SecretString != nil {
		return *result.SecretString, nil
	}
	return "", fmt.Errorf("secret string is empty")
}

// getTranscript uses youtube-transcript-api-go to fetch subtitles
func getTranscript(videoID string) (string, error) {
	client := yt_transcript.NewClient()
	// Try Japanese first
	text, err := client.GetFormattedTranscripts(videoID, []string{"ja"}, false)
	if err != nil {
		return "", fmt.Errorf("failed to get transcript: %w", err)
	}
	return text, nil
}

func generateSummary(ctx context.Context, transcript, title string) (*SummaryData, error) {
	// Truncate transcript to avoid token limits
	if len(transcript) > 20000 {
		transcript = transcript[:20000]
	}

	prompt := fmt.Sprintf(`以下のYouTube動画の字幕テキストを元に、以下の2種類の要約をJSON形式で出力してください。

1. short_summary: 400文字程度の簡潔な要約（動画を見るかどうか判断できる情報を含める）
2. detail_summary: 4000文字程度の詳細な要約（動画の内容を詳細に解説し、視聴しなくても内容が分かるレベルにする。章立てや箇条書き（Markdown形式）を使って読みやすくすること）

動画タイトル: %s

字幕テキスト:
%s

出力形式（必ずこのJSONフォーマットのみを出力してください）:
{
  "short_summary": "...",
  "detail_summary": "..."
}`, title, transcript)

	// Claude Messages API request body
	reqBody := map[string]interface{}{
		"anthropic_version": "bedrock-2023-05-31",
		"max_tokens":        4096, // Increased tokens for detailed summary
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := bedrockClient.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String("arn:aws:bedrock:us-east-1:031921999648:inference-profile/global.anthropic.claude-haiku-4-5-20251001-v1:0"),
		ContentType: aws.String("application/json"),
		Body:        reqJSON,
	})
	if err != nil {
		return nil, fmt.Errorf("bedrock invoke failed: %w", err)
	}

	// Parse response from Bedrock
	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse bedrock response: %w", err)
	}

	if len(result.Content) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	responseText := result.Content[0].Text

	// Clean up response text (remove markdown code blocks if present)
	responseText = strings.TrimSpace(responseText)
	if strings.Contains(responseText, "```") {
		// Remove ```json and ``` 
		responseText = strings.ReplaceAll(responseText, "```json", "")
		responseText = strings.ReplaceAll(responseText, "```", "")
		responseText = strings.TrimSpace(responseText)
	}

	// Parse JSON output from Claude
	var summaryData SummaryData
	if err := json.Unmarshal([]byte(responseText), &summaryData); err != nil {
		// Fallback: try to extract JSON if Claude added text usually shouldn't happen with strict prompt
		return nil, fmt.Errorf("failed to parse summary json: %w. Response: %s", err, responseText)
	}

	return &summaryData, nil
}

func getVideoItem(ctx context.Context, videoID string) (map[string]types.AttributeValue, error) {
	// Query GSI to get Primary Key
	input := &dynamodb.QueryInput{
		TableName:              aws.String(tableName),
		IndexName:              aws.String("videoId-index"),
		KeyConditionExpression: aws.String("videoId = :vid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":vid": &types.AttributeValueMemberS{Value: videoID},
		},
		Limit: aws.Int32(1),
	}

	resp, err := dynamoClient.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query index: %w", err)
	}

	if len(resp.Items) == 0 {
		return nil, nil // Not found
	}

	// Found in index. Get full item from base table.
	item := resp.Items[0]
	hashtag := item["hashtag"].(*types.AttributeValueMemberS).Value
	processedAt := item["processedAt"].(*types.AttributeValueMemberS).Value

	getItemInput := &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"hashtag":     &types.AttributeValueMemberS{Value: hashtag},
			"processedAt": &types.AttributeValueMemberS{Value: processedAt},
		},
	}

	getResult, err := dynamoClient.GetItem(ctx, getItemInput)
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	return getResult.Item, nil
}

func saveVideoData(ctx context.Context, channelID string, video VideoDetails, transcript string, summary *SummaryData) error {
	log.Printf("DEBUG: Saving video data for %s to table %s", video.ID, tableName)
	now := time.Now().UTC().Format(time.RFC3339)

	// Flatten thumbnails to a map if needed, or store as Map/JSON
	// For simplicity, we just store the medium URL
	thumbURL := ""
	if video.Thumbnails != nil && video.Thumbnails.Medium != nil {
		thumbURL = video.Thumbnails.Medium.Url
	}

	item := map[string]types.AttributeValue{
		"hashtag":      &types.AttributeValueMemberS{Value: channelID}, // Using "hashtag" key for PK compatibility
		"processedAt":  &types.AttributeValueMemberS{Value: now},
		"videoId":      &types.AttributeValueMemberS{Value: video.ID},
		"title":        &types.AttributeValueMemberS{Value: video.Title},
		"channelTitle": &types.AttributeValueMemberS{Value: video.ChannelTitle},
		"publishedAt":  &types.AttributeValueMemberS{Value: video.PublishedAt},
		"viewCount":    &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", video.ViewCount)},
		"likeCount":    &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", video.LikeCount)},
		"thumbnailUrl": &types.AttributeValueMemberS{Value: thumbURL},
		"transcript":   &types.AttributeValueMemberS{Value: transcript},
	}

	if summary != nil {
		item["summary"] = &types.AttributeValueMemberS{Value: summary.ShortSummary}
		item["detailSummary"] = &types.AttributeValueMemberS{Value: summary.DetailSummary}
	}

	_, err := dynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	})
	return err
}

func handler(ctx context.Context) (BatchStats, error) {
	log.Println("Starting batch processing (Go) - Channel mode (Search.List)")

	stats := BatchStats{}
	tableName = os.Getenv("DYNAMODB_TABLE")
	if tableName == "" {
		tableName = "youtube-summary-dev"
	}

	// Get channel ID from environment
	channelID := os.Getenv("CHANNEL_ID")
	if channelID == "" {
		// For backward compatibility or testing defaults
		channelID = "UC2kM01yXNnouBsJJ0ghyfMg"
	}
	log.Printf("Processing channel: %s", channelID)

	// Get Secrets
	ytSecret := os.Getenv("YOUTUBE_API_SECRET")
	if ytSecret == "" {
		ytSecret = "youtube-summary/youtube-api-key"
	}

	ytKey, err := getSecret(ctx, ytSecret)
	if err != nil {
		log.Printf("Error getting YouTube API key: %v", err)
		return stats, err
	}

	// YouTube Client
	ytService, err := youtube.NewService(ctx, option.WithAPIKey(ytKey))
	if err != nil {
		log.Printf("Error creating YouTube service: %v", err)
		return stats, err
	}

	// 1. Search for recent videos (including live archives)
	// We use Search.List with order=date to get the latest videos.
	searchCall := ytService.Search.List([]string{"id"}).
		ChannelId(channelID).
		Order("date").
		Type("video").
		MaxResults(50) // 50 is the maximum allowed by YouTube API per request

	searchResp, err := searchCall.Do()
	if err != nil {
		return stats, fmt.Errorf("error searching videos: %w", err)
	}

	stats.VideosFound = len(searchResp.Items)
	log.Printf("Found %d videos in search results", stats.VideosFound)

	if len(searchResp.Items) == 0 {
		return stats, nil
	}

	// Collect Video IDs
	var videoIDs []string
	for _, item := range searchResp.Items {
		videoIDs = append(videoIDs, item.Id.VideoId)
	}

	// 2. Get Video Details (Stats, ContentDetails)
	// Search API doesn't return viewCount or likeCount, so we need Videos.List
	videosCall := ytService.Videos.List([]string{"snippet", "statistics", "contentDetails"}).
		Id(strings.Join(videoIDs, ","))
	
	videosResp, err := videosCall.Do()
	if err != nil {
		return stats, fmt.Errorf("error fetching video details: %w", err)
	}

	for _, item := range videosResp.Items {
		videoID := item.Id
		title := item.Snippet.Title
		log.Printf("Processing video: %s (%s)", title, videoID)

		// Create VideoDetails struct for saving later
		videoDetails := VideoDetails{
			ID:           videoID,
			Title:        title,
			PublishedAt:  item.Snippet.PublishedAt,
			ChannelTitle: item.Snippet.ChannelTitle,
			ViewCount:    item.Statistics.ViewCount,
			LikeCount:    item.Statistics.LikeCount,
			Thumbnails: &youtube.ThumbnailDetails{
				Medium: item.Snippet.Thumbnails.Medium,
			},
		}
		if item.Snippet.Thumbnails != nil {
			videoDetails.Thumbnails = item.Snippet.Thumbnails
		}

		// Check if already processed
		existingItem, err := getVideoItem(ctx, videoID)
		if err != nil {
			log.Printf("Error checking DB for %s: %v", videoID, err)
			// Continue or fail? Continue trying to process seems safe.
		}

		if existingItem != nil {
			// Check if detailSummary exists
			if _, ok := existingItem["detailSummary"]; ok {
				log.Printf("Video %s already has summary. Skipping.", videoID)
				stats.VideosAlreadyProcessed++
				continue
			}
		}

		// Retrieve or Fetch Transcript
		var transcript string
		// Check DB first
		if existingItem != nil {
			if t, ok := existingItem["transcript"]; ok {
				transcript = t.(*types.AttributeValueMemberS).Value
				log.Printf("Found existing transcript for %s", videoID)
			}
		}

		// If no transcript, and LOCAL_RUN, fetch it
		if transcript == "" {
			if os.Getenv("LOCAL_RUN") == "true" {
				// Check filters (optional)
				vc := item.Statistics.ViewCount
				if vc < 100 {
					log.Printf("Notice: Video %s has low views (%d)", videoID, vc)
				}

				log.Printf("Fetching transcript for %s...", videoID)
				fetchedTx, err := getTranscript(videoID)
				if err != nil {
					log.Printf("No transcript found for %s: %v", videoID, err)
					stats.VideosWithoutTx++
					continue
				}
				transcript = fetchedTx

				// Save transcript immediately to avoid re-fetching
				if err := saveVideoData(ctx, channelID, videoDetails, transcript, nil); err != nil {
					log.Printf("Error saving transcript for %s: %v", videoID, err)
					// Proceed anyway to try summarizing?
				} else {
					log.Printf("Saved transcript for %s", videoID)
				}

				// Add delay to avoid YouTube rate limiting locally
				time.Sleep(3 * time.Second)

			} else {
				// AWS Lambda mode but no transcript in DB
				log.Printf("Skipping video %s: No transcript in DB and not running locally", videoID)
				continue
			}
		}

		// At this point we have a transcript (or we continued).
		// Generate summary with Bedrock
		log.Printf("Generating summary for %s...", videoID)
		summaryData, err := generateSummary(ctx, transcript, title)
		if err != nil {
			log.Printf("Error summarizing %s: %v", videoID, err)
			stats.Errors++
			continue
		}

		// Save processing result (Summary + Transcript + Metadata)
		if err := saveVideoData(ctx, channelID, videoDetails, transcript, summaryData); err != nil {
			log.Printf("Error saving summary for %s: %v", videoID, err)
			stats.Errors++
		} else {
			log.Printf("Successfully processed video %s", videoID)
			stats.VideosSummarized++
		}
	}
	return stats, nil
}

func main() {
	if os.Getenv("LOCAL_RUN") == "true" {
		log.Println("Running in local mode...")
		stats, err := handler(context.Background())
		if err != nil {
			log.Fatalf("Local execution failed: %v", err)
		}
		log.Printf("Local execution finished. Stats: %+v", stats)
	} else {
		lambda.Start(handler)
	}
}
