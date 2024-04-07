package models

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hermes/services/Temporal/TemporalJobs"
	"net/http"
	"time"

	"github.com/tryamigo/themis"
)

type OAuthTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type Instagram struct {
	CreatedAt                           time.Time         `json:"createdAt" bson:"createdAt"`
	UpdatedAt                           time.Time         `json:"updatedAt" bson:"updatedAt"`
	Biography                           string            `json:"biography"`
	IgID                                int64             `json:"ig_id"`
	FollowersCount                      int               `json:"followers_count"`
	FollowsCount                        int               `json:"follows_count"`
	MediaCount                          int               `json:"media_count"`
	Name                                string            `json:"name"`
	ProfilePictureURL                   string            `json:"profile_picture_url"`
	Username                            string            `json:"username"`
	Website                             string            `json:"website"`
	IsConnected                         bool              `json:"isConnected" bson:"isConnected"`
	ConnectedFbPageID                   string            `json:"connectedFbPageID" bson:"connectedFbPageID"`
	ConnectedInstagramBusinessAccountID string            `json:"connectedInstagramBusinessAccountID" bson:"connectedInstagramBusinessAccountID"`
	ConnectedInstagramPage              InstagramPage     `json:"instagramPage" bson:"instagramPage"`
	Insights                            InstagramInsights `json:"insights" bson:"insights"`
	Approved                            bool              `json:"approved" bson:"approved"`
	IsVerified                          bool              `json:"isVerified" bson:"isVerified"`
}

type InstagramInsightsResponse struct {
	Data   []MediaInsight `json:"data"`
	Paging struct {
		Cursors struct {
			After string `json:"after,omitempty"`
		} `json:"cursors"`
	} `json:"paging"`
}

type MediaInsight struct {
	ID            string `json:"id"`
	CommentsCount int    `json:"comments_count"`
	LikeCount     int    `json:"like_count"`
	Type          string `json:"media_type"`
}

type InstagramInsights struct {
	EngagementRate       float64 `json:"engagementRate" bson:"engagementRate"`
	AvgLikes             int64   `json:"avgLikes" bson:"avgLikes"`
	AvgComments          int64   `json:"avgComments" bson:"avgComments"`
	AvgReach             int64   `json:"avgReach" bson:"avgReach"`
	AvgVideoReach        int64   `json:"avgVideoReach" bson:"avgVideoReach"`
	LikesToCommentsRatio int64   `json:"likesToCommentsRatio" bson:"likesToCommentsRatio"`
	TotalLikes           int64   `json:"totalLikes" bson:"totalLikes"`
	Engagement           int     `json:"engagement"`
	Impressions          int     `json:"impressions"`
	Reach                int     `json:"reach"`
	Video_views          int     `json:"video_views"`
	Saved                int     `json:"saved"`
	VideoCount           int     `json:"videoCount"`
}

type MediaWiseInsight struct {
	Engagement  int `json:"engagement"`
	Impressions int `json:"impressions"`
	Reach       int `json:"reach"`
	Video_views int `json:"video_views"`
	Saved       int `json:"saved"`
	VideoCount  int `json:"videoCount"`
}

type MediaWiseInsightResponse struct {
	Data []struct {
		Name   string `json:"name"`
		Values []struct {
			Value int `json:"value"`
		} `json:"values"`
	} `json:"data"`
}

type InstagramOAuthCallback struct {
	AccessToken string `json:"access_token"`
	UserID      string `json:"user_id"`
}

type InstagramPagesResponse struct {
	Data   []InstagramPage `json:"data"`
	Paging struct {
		Cursors struct {
			After string `json:"after,omitempty"`
		} `json:"cursors"`
	} `json:"paging"`
}

type InstagramPage struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
}

type InstagramBusinessAccount struct {
	InstagramAccount struct {
		ID string `json:"id"`
	} `json:"instagram_business_account"`
	ID string `json:"id"`
}

type Snapchat struct {
	ID          string    `json:"id" bson:"_id"`
	CreatedAt   time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt" bson:"updatedAt"`
	IsConnected bool      `json:"isConnected" bson:"isConnected"`
	Approved    bool      `json:"approved" bson:"approved"`
	IsVerified  bool      `json:"isVerified" bson:"isVerified"`
}

func GetInstagramOAuthURL(state string) string {
	url := instagramOAuthConfig.AuthCodeURL(state)
	return url
}

func HandleInstagramCallback(code string) (OAuthTokenResponse, error) {

	token, err := instagramOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		return OAuthTokenResponse{}, fmt.Errorf("failed to exchange token: " + err.Error())
	}

	// Exchange the short-lived access token for a long-lived access token
	longLivedToken, err := exchangeToken(token.AccessToken)
	if err != nil {
		return OAuthTokenResponse{}, fmt.Errorf("failed to exchange short-lived token for long-lived token: " + err.Error())
	}

	return longLivedToken, nil
}

func exchangeToken(shortLivedToken string) (OAuthTokenResponse, error) {

	var tokenResp OAuthTokenResponse
	url := fmt.Sprintf("%s?grant_type=fb_exchange_token&client_id=%s&client_secret=%s&fb_exchange_token=%s", instagramOAuthConfig.Endpoint.TokenURL, InstagramClientID, InstagramClientSecret, shortLivedToken)

	status, resp, err := themis.HitAPIEndpoint2(url, http.MethodGet, nil, nil, nil)
	if err != nil {
		return tokenResp, err
	} else if status >= 400 {
		return tokenResp, errors.New(string(resp))
	}

	err = json.Unmarshal(resp, &tokenResp)
	if err != nil {
		return tokenResp, err
	}

	if len(tokenResp.AccessToken) == 0 {
		return tokenResp, fmt.Errorf("access token is empty, tokenResp was %s", string(resp))
	}

	return tokenResp, nil
}

func GetInstagramUserProfile(accessToken string, influencerID string) (Instagram, error) {
	var profile Instagram

	accountPages, err := GetAssociatedAccountPages(accessToken)
	if err != nil {
		return profile, fmt.Errorf("failed to get associated account pages " + err.Error())
	}

	if len(accountPages) == 0 {
		return profile, fmt.Errorf("no associated account pages found")
	}

	businessAccountID, err := GetInstagramBusinessAccountIDForPage(accountPages[0].ID, accessToken)
	if err != nil {
		return profile, fmt.Errorf("failed to get associated instagram account id " + err.Error())
	}

	profile, err = GetInstagramUserInsights(businessAccountID, accessToken)
	if err != nil {
		return profile, fmt.Errorf("failed to get associated instagram account details " + err.Error())
	}

	// insights, err := GetInstagramAccountInsights(businessAccountID, accessToken, profile.FollowersCount)
	// if err != nil {
	// 	return profile, fmt.Errorf("failed to get associated instagram account insights " + err.Error())
	// }

	profile.CreatedAt = time.Now()
	profile.UpdatedAt = time.Now()
	profile.ConnectedFbPageID = accountPages[0].ID
	profile.ConnectedInstagramPage = accountPages[0]
	profile.ConnectedInstagramBusinessAccountID = businessAccountID

	// Create Instagram Insight Workflow
	err = TemporalJobs.CreateInstagramInsightsFetchWorkflow(businessAccountID, profile.FollowersCount, influencerID)
	if err != nil {
		return profile, fmt.Errorf("instagram profile update error: %s", err.Error())
	}

	// profile.Insights = insights

	return profile, nil
}

func GetAssociatedAccountPages(accessToken string) ([]InstagramPage, error) {

	var response []InstagramPage

	url := "https://graph.facebook.com/v16.0/me/accounts"
	params := [][]string{{"access_token", accessToken}}

	var responseBody InstagramPagesResponse
	status, resp, err := themis.HitAPIEndpoint2(url, http.MethodGet, nil, nil, params)
	if err != nil {
		return response, err
	} else if status >= 400 {
		return response, errors.New(string(resp))
	}

	err = json.Unmarshal(resp, &responseBody)
	if err != nil {
		return response, err
	}

	response = append(response, responseBody.Data...)

	for len(responseBody.Paging.Cursors.After) != 0 {

		params := [][]string{{"after", responseBody.Paging.Cursors.After}, {"access_token", accessToken}}
		status, resp, err := themis.HitAPIEndpoint2(url, http.MethodGet, nil, nil, params)
		if err != nil {
			return response, err
		} else if status >= 400 {
			return response, errors.New(string(resp))
		}

		// Setting the pagination cursor to default value
		responseBody.Paging.Cursors.After = ""

		err = json.Unmarshal(resp, &responseBody)
		if err != nil {
			return response, err
		}

		response = append(response, responseBody.Data...)

	}

	return response, nil
}

func GetInstagramBusinessAccountIDForPage(pageID string, accessToken string) (string, error) {

	url := fmt.Sprintf("https://graph.facebook.com/v16.0/%s?fields=instagram_business_account&access_token=%s", pageID, accessToken)

	var responseBody InstagramBusinessAccount
	status, resp, err := themis.HitAPIEndpoint2(url, http.MethodGet, nil, nil, nil)
	if err != nil {
		return responseBody.InstagramAccount.ID, err
	} else if status >= 400 {
		return responseBody.InstagramAccount.ID, errors.New(string(resp))
	}

	err = json.Unmarshal(resp, &responseBody)
	if err != nil {
		return responseBody.InstagramAccount.ID, err
	}

	if len(responseBody.InstagramAccount.ID) == 0 {
		return responseBody.InstagramAccount.ID, fmt.Errorf("no instagram business acount ID found")
	}

	return responseBody.InstagramAccount.ID, nil
}

func GetInstagramUserInsights(instaAccountID string, accessToken string) (Instagram, error) {

	var profile Instagram

	url := fmt.Sprintf("https://graph.facebook.com/v16.0/%s?access_token=%s&fields=biography,id,ig_id,followers_count,follows_count,media_count,name,profile_picture_url,username,website", instaAccountID, accessToken)

	status, resp, err := themis.HitAPIEndpoint2(url, http.MethodGet, nil, nil, nil)
	if err != nil {
		return profile, err
	} else if status >= 400 {
		return profile, errors.New(string(resp))
	}

	err = json.Unmarshal(resp, &profile)
	if err != nil {
		return profile, err
	}

	return profile, nil
}

func GetInstagramAccountInsights(instaAccountID string, accessToken string, followers int) (InstagramInsights, error) {

	var insights InstagramInsights

	postsCount := 0
	likesCount := 0
	commentsCount := 0

	// media level insight metrics
	engagement := 0
	impressions := 0
	reach := 0
	saved := 0
	video_views := 0
	video_count := 0

	url := fmt.Sprintf("https://graph.facebook.com/v16.0/%s/media", instaAccountID)
	params := [][]string{{"access_token", accessToken}, {"fields", "id,comments_count,like_count,media_type"}}

	status, resp, err := themis.HitAPIEndpoint2(url, http.MethodGet, nil, nil, params)
	if err != nil {
		return insights, err
	} else if status >= 400 {
		return insights, errors.New(string(resp))
	}

	var responseBody InstagramInsightsResponse
	err = json.Unmarshal(resp, &responseBody)
	if err != nil {
		return insights, err
	}

	for _, media := range responseBody.Data {

		mediaWiseInsight, err := GetMediaWiseInsight(media.ID, media.Type, accessToken)
		if err != nil {
			return insights, err
		}

		postsCount += 1
		likesCount += media.LikeCount
		commentsCount += media.CommentsCount

		engagement += mediaWiseInsight.Engagement
		impressions += mediaWiseInsight.Impressions
		reach += mediaWiseInsight.Reach
		saved += mediaWiseInsight.Saved
		video_views += mediaWiseInsight.Video_views
		video_count += mediaWiseInsight.VideoCount
	}

	for len(responseBody.Paging.Cursors.After) != 0 {
		params := [][]string{{"access_token", accessToken}, {"fields", "id,comments_count,like_count,media_type"}, {"after", responseBody.Paging.Cursors.After}}

		status, resp, err := themis.HitAPIEndpoint2(url, http.MethodGet, nil, nil, params)

		if err != nil {
			return insights, err
		} else if status >= 400 {
			return insights, errors.New(string(resp))
		}

		// Setting the pagination cursor to default value
		responseBody.Paging.Cursors.After = ""

		err = json.Unmarshal(resp, &responseBody)
		if err != nil {
			return insights, err
		}

		for _, media := range responseBody.Data {

			mediaWiseInsight, err := GetMediaWiseInsight(media.ID, media.Type, accessToken)
			if err != nil {
				return insights, err
			}

			postsCount += 1
			likesCount += media.LikeCount
			commentsCount += media.CommentsCount

			engagement += mediaWiseInsight.Engagement
			impressions += mediaWiseInsight.Impressions
			reach += mediaWiseInsight.Reach
			saved += mediaWiseInsight.Saved
			video_views += mediaWiseInsight.Video_views
			video_count += mediaWiseInsight.VideoCount
		}
	}

	insights = GetInsightsFromData(postsCount, likesCount, commentsCount, impressions, reach, video_count)

	insights.Engagement = engagement
	insights.Saved = saved
	insights.Video_views = video_views

	return insights, nil
}

func GetInsightsFromData(postsCount int, likesCount int, commentsCount int, impressions int, reach int, video_count int) InstagramInsights {
	return InstagramInsights{
		EngagementRate:       (divideTwoFloat(float64(likesCount+commentsCount), float64(impressions))) * 100,
		AvgLikes:             divideTwoInt(int64(likesCount), int64(postsCount)),
		AvgComments:          divideTwoInt(int64(commentsCount), int64(postsCount)),
		LikesToCommentsRatio: divideTwoInt(int64(likesCount), int64(commentsCount)),
		TotalLikes:           int64(likesCount),
		AvgReach:             divideTwoInt(int64(reach), int64(postsCount)),
		AvgVideoReach:        divideTwoInt(int64(reach), int64(video_count)),
		Reach:                reach,
		Impressions:          impressions,
		VideoCount:           video_count,
	}
}

func divideTwoFloat(a float64, b float64) float64 {
	if b == 0 {
		return 0.0
	}
	return a / b
}

func divideTwoInt(a int64, b int64) int64 {
	if b == 0 {
		return 0
	}
	return a / b
}

func GetMediaWiseInsight(mediaID string, mediaType string, accessToken string) (MediaWiseInsight, error) {
	var mediaWiseInsight MediaWiseInsight

	video_count := 0
	url := fmt.Sprintf("https://graph.facebook.com/v16.0/%s/insights", mediaID)

	var params [][]string

	if mediaType == "VIDEO" {
		params = [][]string{{"access_token", accessToken}, {"metric", "reach,saved,plays,shares"}}
		video_count += 1
	} else {
		params = [][]string{{"access_token", accessToken}, {"metric", "engagement,impressions,reach,saved,video_views"}}
	}

	status, resp, err := themis.HitAPIEndpoint2(url, http.MethodGet, nil, nil, params)
	if err != nil {
		return mediaWiseInsight, err
	} else if status >= 400 {
		return mediaWiseInsight, errors.New(string(resp))
	}

	var responseBody MediaWiseInsightResponse
	err = json.Unmarshal(resp, &responseBody)
	if err != nil {
		return mediaWiseInsight, err
	}

	mediaWiseInsight.Engagement = getMediaWiseInsightKeyValue(responseBody, "engagement")
	mediaWiseInsight.Impressions = getMediaWiseInsightKeyValue(responseBody, "impressions")
	mediaWiseInsight.Reach = getMediaWiseInsightKeyValue(responseBody, "reach")
	mediaWiseInsight.Saved = getMediaWiseInsightKeyValue(responseBody, "saved")
	mediaWiseInsight.Video_views = getMediaWiseInsightKeyValue(responseBody, "video_views")
	mediaWiseInsight.VideoCount = video_count
	return mediaWiseInsight, nil
}

func getMediaWiseInsightKeyValue(mediaInsight MediaWiseInsightResponse, key string) int {

	for _, insight := range mediaInsight.Data {
		if insight.Name == key {
			return insight.Values[0].Value
		}
	}

	return 0
}
