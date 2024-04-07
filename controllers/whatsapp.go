package controllers

import (
	"encoding/json"
	"fmt"
	"hermes/models"
	"hermes/utils/data"
	"net/http"
	"os"
)

type User struct {
	WaID      string `json:"waId"`
	WaNumber  string `json:"waNumber"`
	WaName    string `json:"waName"`
	Timestamp string `json:"timestamp"`
}

type AuthLinkResponse struct {
	Ok    bool   `json:"ok"`
	User  User   `json:"user"`
	Error string `json:"error"`
}

func getDataFromWhatsappAuthlink(WhatsappID string) (AuthLinkResponse, error) {
	var response AuthLinkResponse
	var postData = map[string]string{
		"waId": WhatsappID,
	}
	// make a request to the authlink server
	// and return the data
	var err error
	req, err := http.NewRequest("POST", "https://amigo.authlink.me", data.InterfaceToIoReader(postData))
	if err != nil {
		return response, err
	}

	// add headers to the request
	req.Header.Add("clientId", os.Getenv("AUTHLINK_CLIENT_ID"))
	req.Header.Add("clientSecret", os.Getenv("AUTHLINK_CLIENT_SECRET"))
	req.Header.Add("Content-Type", "application/json")

	// make the request using the default client
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return response, err
	}

	// print the response body
	fmt.Printf("resp.Body: %+v\n", (resp.Body))

	// read the response body and unmarshal it into the response struct
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return response, err
	}

	fmt.Printf("response: %v\n", response)

	return response, nil
}

func CreateUserFromAuthLinkResponse(authLinkResponse AuthLinkResponse) (models.User, error) {
	var user models.User
	user.ID = data.GetUUIDString("user")
	user.Phone = authLinkResponse.User.WaNumber
	user.Name = authLinkResponse.User.WaName
	_, err := user.Create()
	if err != nil {
		return user, err
	}
	return user, nil
}
