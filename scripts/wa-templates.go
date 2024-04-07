package scripts

import (
	"hermes/utils/whatsapp"
)

func templateTest(phone string) (err error) {
	variables, err := whatsapp.ParseJSONForMPM("testtemplate")
	if err != nil {
		return
	}
	// fmt.Printf("variables: %v\n", variables)
	err = whatsapp.SendMessageWithTemplate2(phone, "testtemplate", variables, whatsapp.Unknown, nil)
	return
}
