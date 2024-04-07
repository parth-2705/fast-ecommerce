package models

import (
	"strings"

	"github.com/tryamigo/themis"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var InstagramClientID string
var InstagramClientSecret string
var InstagramRedirectURI string
var InstagramOAuthScopes string

var YouTubeClientID string
var YouTubeClientSecret string
var YouTubeRedirectURI string
var YouTubeOAuthScopes string

var (
	instagramOAuthConfig *oauth2.Config
	youtubeOAuthConfig   *oauth2.Config
)

func SocialInit() error {

	err := InstagramInit()
	if err != nil {
		return err
	}

	err = YouTubeInit()
	if err != nil {
		return err
	}

	return nil
}

func InstagramInit() error {
	instagramClientID, err := themis.GetSecret("INSTAGRAM_CLIENT_ID")
	if err != nil {
		return err
	}
	instagramClientSecret, err := themis.GetSecret("INSTAGRAM_CLIENT_SECRET")
	if err != nil {
		return err
	}
	instagramRedirectURI, err := themis.GetSecret("INSTAGRAM_REDIRECT_URI")
	if err != nil {
		return err
	}
	instagramOAuthScopes, err := themis.GetSecret("INSTAGRAM_OAUTH_SCOPES")
	if err != nil {
		return err
	}

	InstagramClientID = instagramClientID
	InstagramClientSecret = instagramClientSecret
	InstagramRedirectURI = instagramRedirectURI
	InstagramOAuthScopes = instagramOAuthScopes

	instagramOAuthConfig = &oauth2.Config{
		ClientID:     instagramClientID,
		ClientSecret: instagramClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.facebook.com/v16.0/dialog/oauth",
			TokenURL: "https://graph.facebook.com/v16.0/oauth/access_token",
		},
		RedirectURL: instagramRedirectURI,
		Scopes:      strings.Split(instagramOAuthScopes, ","),
	}

	return nil
}

func YouTubeInit() error {
	youtubeClientID, err := themis.GetSecret("YOUTUBE_CLIENT_ID")
	if err != nil {
		return err
	}

	youtubeClientSecret, err := themis.GetSecret("YOUTUBE_CLIENT_SECRET")
	if err != nil {
		return err
	}

	youtubeRedirectURI, err := themis.GetSecret("YOUTUBE_REDIRECT_URI")
	if err != nil {
		return err
	}

	youtubeOAuthScopes, err := themis.GetSecret("YOUTUBE_OAUTH_SCOPES")
	if err != nil {
		return err
	}

	YouTubeClientID = youtubeClientID
	YouTubeClientSecret = youtubeClientSecret
	YouTubeRedirectURI = youtubeRedirectURI
	YouTubeOAuthScopes = youtubeOAuthScopes

	youtubeOAuthConfig = &oauth2.Config{
		ClientID:     YouTubeClientID,
		ClientSecret: YouTubeClientSecret,
		RedirectURL:  YouTubeRedirectURI,
		Scopes:       strings.Split(YouTubeOAuthScopes, ","),
		Endpoint:     google.Endpoint,
	}

	return nil
}
