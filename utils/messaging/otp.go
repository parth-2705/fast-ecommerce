package messaging

import (
	"fmt"
	"os"

	"github.com/twilio/twilio-go"
	verify "github.com/twilio/twilio-go/rest/verify/v2"
)

var twilioClient *twilio.RestClient

func Init() {
	twilioClient = twilio.NewRestClient()
}

func SendPhoneOTP(phoneNumber string) error {
	params := &verify.CreateVerificationParams{}
	params.SetTo(phoneNumber)
	params.SetChannel("sms")
	params.SetTemplateSid(os.Getenv("TWILIO_VERIFICATION_TEMPLATE_SID"))

	resp, err := twilioClient.VerifyV2.CreateVerification(os.Getenv("TWILIO_VERIFICATION_SID"), params)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		if resp.Status != nil {
			fmt.Println(*resp.Status)
		} else {
			fmt.Println("Got nil status")
		}
	}
	return err
}

type OTPResponse struct {
	Valid bool `json:"valid"`
}

func CheckPhoneOTP(phoneNumber string, otp string) (bool, string ,error) {
	params := &verify.CreateVerificationCheckParams{}
	params.SetTo(phoneNumber)
	params.SetCode(otp)

	resp, err := twilioClient.VerifyV2.CreateVerificationCheck(os.Getenv("TWILIO_VERIFICATION_SID"), params)
	if err != nil {
		fmt.Println(err.Error())
		return false, "",err
	} else {
		if resp.Status != nil {
			fmt.Println(*resp.Status)
		} else {
			fmt.Println(resp.Status)
		}
	}

	if resp.Valid == nil {
		return false, "",fmt.Errorf("got invalid response from Twillio")
	} 

	return *resp.Valid, *resp.To, nil
}
