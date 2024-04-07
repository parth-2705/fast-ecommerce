package whatsapp

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/tryamigo/themis"
)

type MagicURLGenerateRequest struct {
	Phone string `json:"phone"`
	Path  string `json:"path"`
}

func GetMagicURLFromWhatsapp(phone string, path string) (code string, err error) {
	endpoint := WHATSAPP_SERVICE_BASE_URL + "/magic/get"
	jsonBody := MagicURLGenerateRequest{
		Phone: phone[1:13],
		Path:  path,
	}
	body, _ := json.Marshal(jsonBody)
	headers := [][]string{
		{"Authorization", fmt.Sprintf("Bearer %s", WHATSAPP_SERVICE_API_KEY)},
	}
	status, resp, err := themis.HitAPIEndpoint2(endpoint, "POST", body, headers, nil)
	if err != nil {
		fmt.Println("error in sending request", err)
		return
	}
	response := map[string]string{}
	json.Unmarshal(resp, &response)
	if status > 400 {
		err = errors.New(response["error"])
		return
	}
	code = response["code"]
	return
}
