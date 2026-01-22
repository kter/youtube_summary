package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

var secretsClient *secretsmanager.Client

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-northeast-1"))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	secretsClient = secretsmanager.NewFromConfig(cfg)
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

func getTranscriptDebug(videoID string) (string, error) {
	url := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	body := string(bodyBytes)

    // DEBUG: Check for captionTracks existence
	if !strings.Contains(body, "captionTracks") {
        // Output a snippet of the body to see what's going on (avoid too much log)
        // Check if it's "Sign in" page or something
        if strings.Contains(body, "Sign in to prove") {
            return "", fmt.Errorf("YouTube blocking (Sign in required)")
        }
		return "", fmt.Errorf("no captionTracks found in HTML. Body length: %d", len(body))
	}
    
    fmt.Println("Found 'captionTracks' in HTML!")

	parts := strings.Split(body, `"captionTracks":`)
	if len(parts) < 2 {
		return "", fmt.Errorf("failed to parse captionTracks")
	}
    
    jsonPart := parts[1]
    endIndex := strings.Index(jsonPart, "]")
    if endIndex == -1 {
         return "", fmt.Errorf("failed to find end of captionTracks")
    }
    captionTracksJSON := jsonPart[:endIndex+1]
    fmt.Printf("Extracted JSON: %s\n", captionTracksJSON)

	return "Found", nil
}

func main() {
	ctx := context.Background()
	channelID := "UC2kM01yXNnouBsJJ0ghyfMg" // @noiehoie

	// Get Secrets
	ytSecret := "youtube-summary/youtube-api-key"
	ytKey, err := getSecret(ctx, ytSecret)
	if err != nil {
		log.Fatalf("Error getting key: %v", err)
	}

	ytService, err := youtube.NewService(ctx, option.WithAPIKey(ytKey))
	if err != nil {
		log.Fatalf("Error creating service: %v", err)
	}

	// 1. Try Search.List (Existing Method)
	fmt.Println("=== Testing Search.List ===")
	call := ytService.Search.List([]string{"id", "snippet"}).
		ChannelId(channelID).
		Type("video").
		Order("date").
		MaxResults(20) // Check first 20

	resp, err := call.Do()
	if err != nil {
		log.Fatalf("Search Error: %v", err)
	}
	fmt.Printf("Search found %d videos\n", len(resp.Items))
    for _, item := range resp.Items {
        fmt.Printf("- %s (%s) : %s\n", item.Snippet.Title, item.Id.VideoId, item.Snippet.PublishedAt)
    }

	// 2. Try Channels.List + PlaylistItems (Proposed Method)
    fmt.Println("\n=== Testing PlaylistItems (Uploads) ===")
    chanCall := ytService.Channels.List([]string{"contentDetails"}).Id(channelID)
    chanResp, err := chanCall.Do()
    if err != nil {
        log.Fatalf("Channels Error: %v", err)
    }
    if len(chanResp.Items) == 0 {
        log.Fatal("Channel not found")
    }
    uploadsID := chanResp.Items[0].ContentDetails.RelatedPlaylists.Uploads
    fmt.Printf("Uploads Playlist ID: %s\n", uploadsID)

    plCall := ytService.PlaylistItems.List([]string{"snippet"}).
        PlaylistId(uploadsID).
        MaxResults(20)
    
    plResp, err := plCall.Do()
    if err != nil {
        log.Fatalf("PlaylistItems Error: %v", err)
    }
    fmt.Printf("Playlist found %d videos\n", len(plResp.Items))
    
    var targetVideoID string
    for _, item := range plResp.Items {
        fmt.Printf("- %s (%s) : %s\n", item.Snippet.Title, item.Snippet.ResourceId.VideoId, item.Snippet.PublishedAt)
        if targetVideoID == "" {
            targetVideoID = item.Snippet.ResourceId.VideoId
        }
    }

    // 3. Debug Transcript Logic
    if targetVideoID != "" {
        fmt.Printf("\n=== Testing Transcript for latest video: %s ===\n", targetVideoID)
        _, err := getTranscriptDebug(targetVideoID)
        if err != nil {
            fmt.Printf("Transcript Error: %v\n", err)
        } else {
            fmt.Println("Transcript extraction logic passed initial check.")
        }
    }
}
