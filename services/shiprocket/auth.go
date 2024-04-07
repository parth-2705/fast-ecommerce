package shiprocket

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/tryamigo/themis"
)

type ShipRocketAuthRequest struct {
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
}

type ShipRocketAuthResponse struct {
	ID        int    `json:"id,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Email     string `json:"email,omitempty"`
	CompanyID int    `json:"company_id,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	Token     string `json:"token,omitempty"`
}

func GetShipRocketToken() (string, error) {

	email, err := themis.GetSecret("SHIPROCKET_EMAIL")
	if err != nil {
		panic(err)
	}

	password, err := themis.GetSecret("SHIPROCKET_PASSWORD")
	if err != nil {
		panic(err)
	}

	shipRocketAuthRequest := ShipRocketAuthRequest{
		Email:    email,
		Password: password,
	}

	requestBody, _ := json.Marshal(shipRocketAuthRequest)

	url := "https://apiv2.shiprocket.in/v1/external/auth/login"
	statusCode, responseBody, err := themis.HitAPIEndpoint2(url, http.MethodPost, requestBody, [][]string{}, [][]string{})
	if err != nil {
		return "", err
	}
	if statusCode >= 400 {
		return "", fmt.Errorf("erorr: %s with status code: %d", string(responseBody), statusCode)
	}

	var shipRocketAuthResponse ShipRocketAuthResponse
	err = json.Unmarshal(responseBody, &shipRocketAuthResponse)
	if err != nil {
		return "", err
	}

	return shipRocketAuthResponse.Token, nil
}
