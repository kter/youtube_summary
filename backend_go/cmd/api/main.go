package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"

	"sort"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var (
	dynamoClient *dynamodb.Client
	tableName    string
)

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-northeast-1"))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	dynamoClient = dynamodb.NewFromConfig(cfg)
}

func createResponse(statusCode int, body interface{}) (events.APIGatewayV2HTTPResponse, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return events.APIGatewayV2HTTPResponse{StatusCode: 500}, err
	}
	return events.APIGatewayV2HTTPResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
            "Access-Control-Allow-Origin": "*",
		},
		Body: string(jsonBody),
	}, nil
}

func getSummaries(ctx context.Context, channelID string, limit int) ([]map[string]interface{}, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(tableName),
		KeyConditionExpression: aws.String("hashtag = :h"), // Using "hashtag" as PK name for compatibility
		// Sort by processedAt descending
		ScanIndexForward: aws.Bool(false),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":h": &types.AttributeValueMemberS{Value: channelID},
		},
		Limit: aws.Int32(int32(limit)),
	}

	resp, err := dynamoClient.Query(ctx, input)
	if err != nil {
		return nil, err
	}

	summaries := []map[string]interface{}{}
	for _, item := range resp.Items {
		// Manual Unmarshal to simple map
		summary := make(map[string]interface{})
		// Helper to unmarshal string/number fields
		if v, ok := item["videoId"].(*types.AttributeValueMemberS); ok {
			summary["videoId"] = v.Value
		}
		if v, ok := item["title"].(*types.AttributeValueMemberS); ok {
			summary["title"] = v.Value
		}
		if v, ok := item["summary"].(*types.AttributeValueMemberS); ok {
			summary["summary"] = v.Value
		}
		if v, ok := item["detailSummary"].(*types.AttributeValueMemberS); ok {
			summary["detailSummary"] = v.Value
		}
		if v, ok := item["processedAt"].(*types.AttributeValueMemberS); ok {
			summary["processedAt"] = v.Value
		}
		if v, ok := item["publishedAt"].(*types.AttributeValueMemberS); ok {
			summary["publishedAt"] = v.Value
		}
		if v, ok := item["channelTitle"].(*types.AttributeValueMemberS); ok {
			summary["channelTitle"] = v.Value
		}
		if v, ok := item["viewCount"].(*types.AttributeValueMemberN); ok {
			summary["viewCount"] = v.Value
		}
		if v, ok := item["likeCount"].(*types.AttributeValueMemberN); ok {
			summary["likeCount"] = v.Value
		}
		if v, ok := item["thumbnailUrl"].(*types.AttributeValueMemberS); ok {
			summary["thumbnails"] = map[string]interface{}{
				"medium": map[string]string{"url": v.Value},
			}
		} else {
			if vMap, ok := item["thumbnails"].(*types.AttributeValueMemberM); ok {
				summary["thumbnails"] = vMap.Value
			}
		}

		summaries = append(summaries, summary)
	}
	// Sort summaries by publishedAt descending (newest first)
	sort.Slice(summaries, func(i, j int) bool {
		p1, _ := summaries[i]["publishedAt"].(string)
		p2, _ := summaries[j]["publishedAt"].(string)
		// Descending order (newest first)
		return p1 > p2
	})

	return summaries, nil
}

func handler(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	tableName = os.Getenv("DYNAMODB_TABLE")
	if tableName == "" {
		tableName = "youtube-summary-dev"
	}

	// Get channel ID from environment
	channelID := os.Getenv("CHANNEL_ID")
	if channelID == "" {
		channelID = "UC2kM01yXNnouBsJJ0ghyfMg" // @noiehoie default
	}

	path := request.RawPath

	if path == "/api/summaries" {
		limitStr := request.QueryStringParameters["limit"]
		limit := 50
		if limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil {
				limit = l
			}
		}

		summaries, err := getSummaries(ctx, channelID, limit)
		if err != nil {
			log.Printf("Error getting summaries: %v", err)
			return createResponse(500, map[string]string{"error": "Internal server error"})
		}

		return createResponse(200, map[string]interface{}{
			"channelId": channelID,
			"count":     len(summaries),
			"summaries": summaries,
		})
	}

	return createResponse(404, map[string]string{"error": "Not Found"})
}

func main() {
	lambda.Start(handler)
}
