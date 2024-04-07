package whatsapp

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/getsentry/sentry-go"
	"github.com/tryamigo/themis"
)

func MarkOrderCompletedOnWhatsapp(phone string) (err error) {

	defer sentry.Recover()

	endpoint := WHATSAPP_SERVICE_BASE_URL + "/order-complete"
	jsonBody := map[string]string{
		"phone": phone[1:13],
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
	return
}
