package models

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hermes/services/Temporal/TemporalJobs"
	"net/http"
	"strconv"
	"time"

	"github.com/tryamigo/themis"
	"golang.org/x/oauth2"
)

type YouTube struct {
	CreatedAt   time.Time       `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt" bson:"updatedAt"`
	IsConnected bool            `json:"isConnected" bson:"isConnected"`
	Insights    YouTubeInsights `json:"insights" bson:"insights"`

	Title       string    `json:"title" bson:"title"`
	Description string    `json:"description" bson:"description"`
	CustomUrl   string    `json:"customUrl" bson:"customUrl"`
	PublishedAt time.Time `json:"publishedAt" bson:"publishedAt"`
	Country     string    `json:"country" bson:"country"`
	Thumbnails  string    `json:"thumbnails" bson:"thumbnails"`
	Approved    bool      `json:"approved" bson:"approved"`
	IsVerified  bool      `json:"isVerified" bson:"isVerified"`
}

type YouTubeInsights struct {
	EngagementRate       float64 `json:"engagementRate" bson:"engagementRate"`
	AvgLikes             int64   `json:"avgLikes" bson:"avgLikes"`
	AvgComments          int64   `json:"avgComments" bson:"avgComments"`
	LikesToCommentsRatio int64   `json:"likesToCommentsRatio" bson:"likesToCommentsRatio"`
	TotalLikes           int64   `json:"totalLikes" bson:"totalLikes"`

	Views                   int64   `json:"views" bson:"views"`
	Comments                int64   `json:"comments" bson:"comments"`
	Dislikes                int64   `json:"dislikes" bson:"dislikes"`
	Shares                  int64   `json:"shares" bson:"shares"`
	EstimatedMinutesWatched int64   `json:"estimatedMinutesWatched" bson:"estimatedMinutesWatched"`
	AverageViewDuration     int64   `json:"averageViewDuration" bson:"averageViewDuration"`
	SubscribersGained       int64   `json:"subscribersGained" bson:"subscribersGained"`
	SubscribersLost         int64   `json:"subscribersLost" bson:"subscribersLost"`
	AverageViewPercentage   float64 `json:"averageViewPercentage" bson:"averageViewPercentage"`
	ViewCount               int     `json:"viewCount" bson:"viewCount"`
	SubscriberCount         int     `json:"subscriberCount" bson:"subscriberCount"`
	VideoCount              int     `json:"videoCount" bson:"videoCount"`
}

type YouTubeChannelStatisticsResponseBody struct {
	Items []struct {
		Kind    string `json:"kind,omitempty"`
		Etag    string `json:"etag,omitempty"`
		ID      string `json:"id,omitempty"`
		Snippet struct {
			Title       string    `json:"title,omitempty"`
			Description string    `json:"description,omitempty"`
			CustomURL   string    `json:"customUrl,omitempty"`
			PublishedAt time.Time `json:"publishedAt,omitempty"`
			Thumbnails  struct {
				Default struct {
					URL    string `json:"url,omitempty"`
					Width  int    `json:"width,omitempty"`
					Height int    `json:"height,omitempty"`
				} `json:"default,omitempty"`
				Medium struct {
					URL    string `json:"url,omitempty"`
					Width  int    `json:"width,omitempty"`
					Height int    `json:"height,omitempty"`
				} `json:"medium,omitempty"`
				High struct {
					URL    string `json:"url,omitempty"`
					Width  int    `json:"width,omitempty"`
					Height int    `json:"height,omitempty"`
				} `json:"high,omitempty"`
			} `json:"thumbnails,omitempty"`
			Localized struct {
				Title       string `json:"title,omitempty"`
				Description string `json:"description,omitempty"`
			} `json:"localized,omitempty"`
			Country string `json:"country,omitempty"`
		} `json:"snippet,omitempty"`
		Statistics struct {
			ViewCount             string `json:"viewCount,omitempty"`
			SubscriberCount       string `json:"subscriberCount,omitempty"`
			HiddenSubscriberCount bool   `json:"hiddenSubscriberCount,omitempty"`
			VideoCount            string `json:"videoCount,omitempty"`
		} `json:"statistics,omitempty"`
	} `json:"items,omitempty"`
}

type YouTubeReportResponseBody struct {
	Kind          string                `json:"kind,omitempty"`
	ColumnHeaders []YouTubeReportColumn `json:"columnHeaders,omitempty"`
	Rows          [][]float64           `json:"rows,omitempty"`
}

type YouTubeReportColumn struct {
	Name       string `json:"name,omitempty"`
	ColumnType string `json:"columnType,omitempty"`
	DataType   string `json:"dataType,omitempty"`
}

func GetYouTubeOAuthURL(state string) string {
	url := youtubeOAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "consent"))
	return url
}

func HandleYouTubeCallback(code string) (OAuthTokenResponse, error) {

	token, err := youtubeOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		return OAuthTokenResponse{}, fmt.Errorf("failed to exchange token: " + err.Error())
	}

	if len(token.AccessToken) == 0 {
		return OAuthTokenResponse{}, fmt.Errorf("access token is empty")
	}

	return OAuthTokenResponse{
		AccessToken: token.AccessToken,
	}, nil
}

func GetYouTubeUserProfile(accessToken string, influencerID string) (YouTube, error) {

	profile, err := GetYouTubeUserProfileUtil(accessToken, influencerID)
	if err != nil {
		return profile, fmt.Errorf("youtube profile create error: %s", err.Error())
	}

	// Create Youtube Insight Workflow
	err = TemporalJobs.CreateYoutubeInsightsFetchWorkflow(influencerID)
	if err != nil {
		return profile, fmt.Errorf("youtube profile create workflow error: %s", err.Error())
	}

	return profile, nil
}

func GetYouTubeUserProfileUtil(accessToken string, influencerID string) (YouTube, error) {
	var profile YouTube

	url := "https://youtube.googleapis.com/youtube/v3/channels"
	params := [][]string{{"part", "snippet,statistics"}, {"mine", "true"}}
	headers := [][]string{{"Authorization", fmt.Sprintf("Bearer %s", accessToken)}}

	var response YouTubeChannelStatisticsResponseBody
	status, resp, err := themis.HitAPIEndpoint2(url, http.MethodGet, nil, headers, params)
	if err != nil {
		return profile, err
	} else if status >= 400 {
		return profile, errors.New(string(resp))
	}

	err = json.Unmarshal(resp, &response)
	if err != nil {
		return profile, err
	}

	if len(response.Items) == 0 {
		return profile, fmt.Errorf("no items in response")
	}

	// Basic channel details
	profile.Title = response.Items[0].Snippet.Title
	profile.Description = response.Items[0].Snippet.Description
	profile.CustomUrl = response.Items[0].Snippet.CustomURL
	profile.PublishedAt = response.Items[0].Snippet.PublishedAt
	profile.Thumbnails = response.Items[0].Snippet.Thumbnails.Default.URL
	profile.Country = response.Items[0].Snippet.Country

	// Channel Statistics
	profile.Insights.ViewCount, err = strconv.Atoi(response.Items[0].Statistics.ViewCount)
	if err != nil {
		return profile, fmt.Errorf("couldn't cast view count due to: " + err.Error())
	}

	profile.Insights.SubscriberCount, err = strconv.Atoi(response.Items[0].Statistics.SubscriberCount)
	if err != nil {
		return profile, fmt.Errorf("couldn't cast subscriber due to: " + err.Error())
	}

	profile.Insights.VideoCount, err = strconv.Atoi(response.Items[0].Statistics.VideoCount)
	if err != nil {
		return profile, fmt.Errorf("couldn't cast video count due to: " + err.Error())
	}

	url = "https://youtubeanalytics.googleapis.com/v2/reports"
	params = [][]string{{"endDate", time.Now().Format("2006-01-02")}, {"startDate", profile.PublishedAt.Format("2006-01-02")}, {"ids", "channel==MINE"}, {"metrics", "views,comments,likes,dislikes,shares,estimatedMinutesWatched,averageViewDuration,subscribersGained,subscribersLost,averageViewPercentage"}}
	headers = [][]string{{"Authorization", fmt.Sprintf("Bearer %s", accessToken)}}

	var responseBody YouTubeReportResponseBody
	status, resp, err = themis.HitAPIEndpoint2(url, http.MethodGet, nil, headers, params)
	if err != nil {
		return profile, err
	} else if status >= 400 {
		return profile, errors.New(string(resp))
	}

	err = json.Unmarshal(resp, &responseBody)
	if err != nil {
		return profile, err
	}

	profile.CreatedAt = time.Now()
	profile.UpdatedAt = time.Now()
	profile.Insights.Views = int64(getColumnValueByKey(responseBody, "views"))
	profile.Insights.Comments = int64(getColumnValueByKey(responseBody, "comments"))
	profile.Insights.Dislikes = int64(getColumnValueByKey(responseBody, "dislikes"))
	profile.Insights.Shares = int64(getColumnValueByKey(responseBody, "shares"))
	profile.Insights.EstimatedMinutesWatched = int64(getColumnValueByKey(responseBody, "estimatedMinutesWatched"))
	profile.Insights.AverageViewDuration = int64(getColumnValueByKey(responseBody, "averageViewDuration"))
	profile.Insights.SubscribersGained = int64(getColumnValueByKey(responseBody, "subscribersGained"))
	profile.Insights.SubscribersLost = int64(getColumnValueByKey(responseBody, "subscribersLost"))
	profile.Insights.AverageViewPercentage = getColumnValueByKey(responseBody, "averageViewPercentage")

	profile.Insights.TotalLikes = int64(getColumnValueByKey(responseBody, "likes"))
	profile.Insights.AvgLikes = divideTwoInt(profile.Insights.TotalLikes, int64(profile.Insights.VideoCount))
	profile.Insights.AvgComments = divideTwoInt(profile.Insights.Comments, int64(profile.Insights.VideoCount))
	profile.Insights.LikesToCommentsRatio = divideTwoInt(profile.Insights.AvgLikes, profile.Insights.AvgComments)
	profile.Insights.EngagementRate = divideTwoFloat(float64(profile.Insights.TotalLikes+profile.Insights.Comments), float64(profile.Insights.Views)) * 100

	return profile, nil
}

func getColumnValueByKey(response YouTubeReportResponseBody, key string) float64 {

	for idx, columnHeader := range response.ColumnHeaders {
		if columnHeader.Name == key {
			return response.Rows[0][idx]
		}
	}

	return 0
}
