package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/tryamigo/themis"
)

type UserInfo struct {
	Email string `json:"email"`
}

func getUserEmail(token string) (email string, err error) {
	headers := [][]string{{"Authorization", "Bearer " + token}}
	status, resp, err := themis.HitAPIEndpoint2("https://www.googleapis.com/oauth2/v2/userinfo", http.MethodGet, nil, headers, nil)
	if err != nil {
		return
	} else if status >= 400 {
		return "", errors.New(string(resp))
	}
	var userInfo UserInfo
	json.Unmarshal(resp, &userInfo)
	email = userInfo.Email
	return
}
