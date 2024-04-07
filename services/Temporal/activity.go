package Temporal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hermes/models"
	"hermes/services/Sentry"
	"hermes/services/Temporal/TemporalShared"
	"hermes/services/shiprocket"
	"hermes/utils/whatsapp"
	"os"
	"strconv"
	"time"

	"github.com/tryamigo/themis"
	"go.temporal.io/sdk/activity"
)

type ACRActivityInput struct {
	CartID string
	// MessageSender func(componentsFiller whatsapp.ACRMessageBody, phoneNumber string) error
	MessageSenderID int
}

type CouponExpiryActivityInput struct {
	CouponID string
}

type InfluencerCommissionActivityInput struct {
	Amount     int64
	CouponCode string
	OrderID    string
}

var messageSenderIDToFuncMap = map[int]func(componentsFiller whatsapp.ACRMessageBody, phoneNumber string) error{
	0: whatsapp.SendACR20MinsMessage1,
	1: whatsapp.SendACR4HrsMessage,
	2: whatsapp.SendACR24HrsMessage,
	3: whatsapp.SendACR3DaysMessage,
}

func SendACREmail(ctx context.Context, activityInput ACRActivityInput) (err error) {

	// Get Cart from DB
	cart, err := models.GetCart(activityInput.CartID)
	if err != nil {
		return err
	}

	// Check if Cart is marked as Completed
	if cart.Status == models.OrderCompleted {
		return nil
	}

	// Get User from DB
	user, err := models.GetUserByID(cart.UserID)
	if err != nil {
		return err
	}

	if user.Phone == "" {
		return nil
	}

	var customerName = "there"
	if user.Name != "" {
		customerName = user.Name
	}

	if user.MarketingCommDisabled {
		return nil
	}

	allowed, reason := cart.ACRAllowed()
	if !allowed {
		return errors.New(reason)
	}

	// product, err := models.GetCompleteProduct(cart.ProductID)
	// if err != nil {
	// 	return err
	// }

	// Send Email
	// err = whatsapp.SendACR20MinsMessage("https://storage.googleapis.com/roovo-images/rawImages/"+string(product.Thumbnail), "there", product.Name, strconv.FormatFloat(product.Price.SellingPrice, 'f', 0, 64), cart.ID, user.Phone, user.MarketingCommDisabled)
	// if err != nil {
	// 	return nil
	// }

	messageBody := whatsapp.ACRMessageBody{
		CartID:             activityInput.CartID,
		ImageLink:          "https://storage.googleapis.com/roovo-images/rawImages/" + string(cart.Items[0].Variant.Thumbnail),
		CustomerName:       customerName,
		ProductName:        cart.Items[0].Product.Name,
		SellingPrice:       strconv.FormatFloat(cart.Items[0].Variant.Price.SellingPrice, 'f', 0, 64),
		MRP:                strconv.FormatFloat(cart.Items[0].Variant.Price.MRP, 'f', 0, 64),
		DiscountAmount:     strconv.FormatFloat(cart.Items[0].Variant.Price.Discount, 'f', 0, 64),
		DiscountPercentage: strconv.FormatFloat(cart.Items[0].Variant.Price.DiscountPercentage, 'f', 0, 64),
	}

	err = messageSenderIDToFuncMap[activityInput.MessageSenderID](messageBody, user.Phone)
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}

	activity.RecordHeartbeat(ctx, fmt.Sprintf("%d sender called", activityInput.MessageSenderID))

	fmt.Println("Whatsapp Message sent")

	return nil
}

// The Case when User has not yet logged in (we do not have User Information)
// The Case when User buys the product

type ReviewActivityInput struct {
	OrderID string
}

type ReorderActivityInput struct {
	OrderID string
}

type componentsBuilder struct {
	ImageLink string
	ProductID string
	BodyParam map[string]string
}

func ExpireCoupon(ctx context.Context, activityInput CouponExpiryActivityInput) (err error) {
	coupon, err := models.GetCouponByID(activityInput.CouponID)
	if err != nil {
		return
	}
	err = coupon.Expire()
	return
}

func createComponentsFromConfig(config models.WATemplate, builder componentsBuilder) []whatsapp.TemplateComponent {

	header := whatsapp.TemplateComponent{
		Type: "header",
		Parameters: []whatsapp.TemplateParameter{
			{
				Type: "image",
				Image: whatsapp.MediaObject{
					Link: builder.ImageLink,
				},
			},
		},
	}

	bodyParameters := []whatsapp.TemplateParameter{}
	for _, param := range config.BodyParameters {
		bodyParameters = append(bodyParameters, whatsapp.TemplateParameter{
			Type: "text",
			Text: builder.BodyParam[param],
		})
	}

	body := whatsapp.TemplateComponent{
		Type:       "body",
		Parameters: bodyParameters,
	}

	button := whatsapp.TemplateComponent{
		Type:    "button",
		SubType: "quick_reply",
		Index:   0,
		Parameters: []whatsapp.TemplateParameter{
			{Type: "payload", Payload: fmt.Sprintf("%d//%s", whatsapp.ReOrderBut, config.ProductID)},
		},
	}

	return append([]whatsapp.TemplateComponent{}, header, body, button)
}

func SendReOrderMessage(ctx context.Context, ai ReorderActivityInput) (err error) {

	// Get Order from DB
	order, err := models.GetOrder(ai.OrderID)
	if err != nil {
		return
	}

	// Confirm that Order wasn't returned

	// get User from DB
	user, err := models.GetUserByID(order.UserID)
	if err != nil {
		return
	}

	customerName, sendMessage := parseUserInfo(user)
	if !sendMessage {
		return nil
	}

	orderedProduct := order.Cart.Items[0].Product
	// Find Config
	config, err := models.GetTemplateConfig(orderedProduct.ID)
	if err != nil {
		return fmt.Errorf("template Config not found for this profuct, %s", err.Error())
	}

	//Create ComponentBuilder

	bodyParams := map[string]string{
		"customerName":        customerName,
		"productName":         order.Cart.Items[0].Product.Name,
		"productSellingPrice": strconv.FormatFloat(order.Cart.Items[0].Variant.Price.SellingPrice, 'f', 0, 64),
		"productMRP":          strconv.FormatFloat(order.Cart.Items[0].Variant.Price.MRP, 'f', 0, 64),
	}

	components := createComponentsFromConfig(config, componentsBuilder{
		ImageLink: config.ImageLink,
		ProductID: config.ProductID,
		BodyParam: bodyParams,
	})

	err = whatsapp.SendProductReOrderMessage2(components, config.TemplateID, user.Phone)
	return
}

func SendReviewMessage(ctx context.Context, ai ReviewActivityInput) (err error) {

	// Get Order from DB
	order, err := models.GetOrder(ai.OrderID)
	if err != nil {
		return
	}

	// Confirm that Order wasn't returned

	// get User from DB
	user, err := models.GetUserByID(order.UserID)
	if err != nil {
		return
	}

	customerName, sendMessage := parseUserInfo(user)
	if !sendMessage {
		return nil
	}

	// Request to send Review Message

	componentsFiller := whatsapp.ReviewMessageBody{
		ImageLink:    "https://storage.googleapis.com/roovo-images/rawImages/" + string(order.Cart.Items[0].Variant.Thumbnail),
		CustomerName: customerName,
		ProductName:  order.Cart.Items[0].Product.Name,
		ProductID:    order.Cart.Items[0].Product.ID,
		BrandName:    order.Cart.Items[0].Brand.Name,
	}

	err = whatsapp.SendProductReviewMessage(componentsFiller, user.ID, user.Phone)
	return
}

func parseUserInfo(user models.User) (customerName string, sendMessage bool) {

	if user.Phone == "" {
		return
	}

	customerName = "there"
	if user.Name != "" {
		customerName = user.Name
	}

	if user.MarketingCommDisabled {
		return
	}

	sendMessage = true
	return

}

func NDRShipmentActivity(ctx context.Context, ai TemporalShared.NDRWorkflowInput) error {

	// Get NDRs from Shiprocket
	ndrs, err := shiprocket.GetAllShiprocketNDRs()
	if err != nil {
		Sentry.SentryCaptureException(err)
		return err
	}

	// For each shipment check whether NDR Template was triggered in last 24 hours
	for _, ndr := range ndrs {
		shipment, err := models.GetShipmentByAWB(ndr.AwbCode)
		if err != nil {
			Sentry.SentryCaptureException(err)
			return err
		}

		err = shipment.TriggerNDRHandling()
		if err != nil {
			Sentry.SentryCaptureException(err)
			return err
		}
	}

	return nil
}

func RoovoPeCreditActivity(ctx context.Context, ai TemporalShared.RoovoPeCreditWorkflowInput) error {

	// Get the order details by orderID
	order, err := models.GetOrder(ai.OrderID)
	if err != nil {
		fmt.Println("err:", err.Error())
		Sentry.SentryCaptureException(err)
		return err
	}

	eligible, err := order.IsEligibleForCashback()
	if err != nil {
		fmt.Println("err:", err.Error())
		Sentry.SentryCaptureException(err)
		return err
	}

	if eligible {
		referringUserID := order.User.ReferredByUser

		// User has signuped by referral code
		if len(referringUserID) > 0 {
			wallet, err := models.GetUserWallet(referringUserID)
			if err != nil {
				fmt.Println("err: GetUserWallet", err.Error())
				Sentry.SentryCaptureException(err)
				return err
			}

			fmt.Println(order.OrderAmount.TotalAmount, 0.05*order.OrderAmount.TotalAmount)
			err = wallet.AddBalance(0.05 * order.OrderAmount.TotalAmount)
			if err != nil {
				fmt.Println("err: AddBalance", err.Error())
				Sentry.SentryCaptureException(err)
				return err
			}
		}
	}

	return nil
}

func CashbackCreditActivity(ctx context.Context, ai TemporalShared.RoovoPeCreditWorkflowInput) error {

	// Get the order details by orderID
	order, err := models.GetOrder(ai.OrderID)
	if err != nil {
		fmt.Println("err:", err.Error())
		Sentry.SentryCaptureException(err)
		return err
	}

	eligible, err := order.IsEligibleForCashback()
	if err != nil {
		fmt.Println("err:", err.Error())
		Sentry.SentryCaptureException(err)
		return err
	}

	if eligible {
		wallet, err := models.GetUserWallet(order.User.ID)
		if err != nil {
			fmt.Println("err: GetUserWallet", err.Error())
			Sentry.SentryCaptureException(err)
			return err
		}

		fmt.Println(order.OrderAmount.TotalAmount, 0.01*order.OrderAmount.TotalAmount)
		err = wallet.AddBalance(0.01 * order.OrderAmount.TotalAmount)
		if err != nil {
			fmt.Println("err: AddBalance", err.Error())
			Sentry.SentryCaptureException(err)
			return err
		}
	}

	return nil
}

func SendAmbassdorMessageActivity(ctx context.Context, ai TemporalShared.EmbassdorRecruitmentINput) error {

	user, err := models.GetUserByID(ai.UserID)
	if err != nil {
		return err
	}

	err = user.SendAmbassdorProgramInvite()
	if err != nil {
		return err
	}

	return nil

}

func SendAmbassdorDeal(ctx context.Context, ai TemporalShared.DealActivityInput) error {

	// Get Ambassdor
	ambassdor, err := models.GetAmbassadorByID(ai.AmbassdorID)
	if err != nil {
		return err
	}

	// Make Photo
	imageLink, err := ambassdor.GetDealImage(ai.ProductID, ai.ChatMessage)
	if err != nil {
		return err
	}

	// ambassdor, err := models.GetAmbassdorByUserID(ambassdorUser.ID)
	// if err != nil {
	// 	return err
	// }

	// if !ambassdor.TutorialSent {
	// Send Message
	err = whatsapp.SendAmbassdorTutorial(ambassdor.Phone, ai.TutorialImageLink, ai.TemplateTutorial, ai.TutorialBodyFiller)
	if err != nil {
		return err
	}
	ambassdor.TutorialSent = true
	// }

	time.Sleep(10 * time.Second)

	err = whatsapp.SendAmbassdorDeal(imageLink, ambassdor.Phone)
	if err != nil {
		return err
	}

	ambassdor.IncreaseDealSentCount()

	return nil
}

func InstagramInsightsActivity(ctx context.Context, ai TemporalShared.InstagramInsightsFetchWorkflowInput) error {

	influencer, err := models.GetInfluencerByID(ai.InfluencerID)
	if err != nil {
		return err
	}

	vaultHandler := themis.NewVaultHandler(os.Getenv("VAULT_URL"), os.Getenv("VAULT_TOKEN"), ai.InfluencerID+"/instagram")
	_, respBody, err1 := vaultHandler.GetSecretsFromVault()
	if err != nil {
		return err1
	}

	var instagramSecret map[string]string
	err = json.Unmarshal(respBody, &instagramSecret)
	if err != nil {
		return err
	}

	access_token := instagramSecret["access_token"]

	insights, err := models.GetInstagramAccountInsights(ai.InstagramID, access_token, ai.Followers)
	if err != nil {
		return err
	}

	influencer.Instagram.Insights = insights
	err = influencer.Update()
	if err != nil {
		return err
	}

	return nil
}

func CreditCommissionToInfluencer(ctx context.Context, ai TemporalShared.InfluencerCommisionWorkflowInput) error {
	application, err := models.GetInfluencercampaignApplicatonByCode(ai.CouponCode)
	if err != nil {
		return err
	}
	influencer, err := models.GetInfluencerByID(application.InfluencerID)
	if err != nil {
		return err
	}
	wallet, err := models.GetUserWallet(influencer.UserID)
	if err != nil {
		return err
	}
	err = wallet.AddBalance(float64(ai.Amount * 5 / 100))
	if err != nil {
		return err
	}
	return nil
}

func YouTubeInsightsActivity(ctx context.Context, ai TemporalShared.YoutubeInsightsFetchWorkflowInput) error {

	influencer, err := models.GetInfluencerByID(ai.InfluencerID)
	if err != nil {
		return err
	}

	vaultHandler := themis.NewVaultHandler(os.Getenv("VAULT_URL"), os.Getenv("VAULT_TOKEN"), ai.InfluencerID+"/youtube")
	_, respBody, err1 := vaultHandler.GetSecretsFromVault()
	if err != nil {
		return err1
	}

	var instagramSecret map[string]string
	err = json.Unmarshal(respBody, &instagramSecret)
	if err != nil {
		return err
	}

	access_token := instagramSecret["access_token"]

	insights, err := models.GetYouTubeUserProfileUtil(access_token, ai.InfluencerID)
	if err != nil {
		return err
	}

	creationTime := influencer.YouTube.CreatedAt
	influencer.YouTube = insights
	influencer.YouTube.CreatedAt = creationTime
	err = influencer.Update()
	if err != nil {
		return err
	}

	return nil
}

func MarkOrderFulfillableActivity(ctx context.Context, ai TemporalShared.FulfillableOrderWorkflowInput) (err error) {
	order, err := models.GetOrder(ai.OrderID)
	if err != nil {
		return
	}
	err = order.SetFullfillable(true)
	return
}
