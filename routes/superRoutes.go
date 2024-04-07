package routes

import (
	"hermes/controllers"
	"hermes/middleware"
	"hermes/services/shiprocket"

	"github.com/gin-gonic/gin"
)

func PublicRoutes(g *gin.RouterGroup) {

	// All public routes
	home := g.Group("/")
	{
		home.GET("/", controllers.HomePage)
		home.GET("/robots.txt", controllers.RobotsTxt)
		home.GET("/about-us", controllers.AboutUsPage)
		home.GET("/contact-us", controllers.ContactUsGetHandler)
		home.POST("/contact-us", controllers.ContactUsPostHandler)
		home.GET("/get-estimated-delivery-time", shiprocket.ServiceabilityController)
		// privacy policy pages
		home.GET("/privacy-policy", controllers.PrivacyPolicyPage)
		home.GET("/terms-and-conditions", controllers.TermsAndConditionsPage)
		home.GET("/return-policy", controllers.ExchangeAndReturnPolicyPage)
		home.GET("/online-registration-policy", controllers.OnlineRegistrationPolicyPage)
		home.GET("/price-and-payment-policy", controllers.PriceAndPaymentPolicyPage)
		home.GET("/inquiry-whatsapp", controllers.InquiryWhatsapp)
		home.GET("/order-whatsapp", controllers.OrderWhatsapp)
		home.GET("/login-whatsapp", controllers.LoginWhatsapp)
		// health check
		home.GET("/health", controllers.HealthCheck)

	}
	signInUp := g.Group("/auth")
	{
		signInUp.GET("/sign-in-up", controllers.SignInUpPage)
		// signInUp.GET("/otp", controllers.GetOtp)
		signInUp.GET("/otp/submit", controllers.SubmitOTP)
		signInUp.POST("/otp", controllers.CheckOTP)
		signInUp.POST("/otp/send", controllers.SendOTP)
		signInUp.GET("/magic/:code", controllers.MagicLinkLogin)
	}

	// Refer routes
	refer := g.Group("/refer")
	{
		refer.GET("/:code", controllers.ReferralSignUp)
	}

	influencer := g.Group("/influencer")
	{
		influencer.GET("", controllers.InfluencerLandingPage)
		influencer.GET("/instagram/oauth/callback", controllers.HandleInstagramOAuthCallback)
		influencer.GET("/youtube/oauth/callback", controllers.HandleYouTubeOAuthCallback)
	}

	product := g.Group("/product")
	{
		product.GET("/:id", controllers.SingleProductPage)
		product.GET("/get-page", controllers.GetProductPage)
	}

	tracking := g.Group("/track-event")
	{
		tracking.POST("", controllers.SendTrackingEvent)
	}

	wishlist := g.Group("/wishlist")
	{
		wishlist.POST("/add", controllers.WishListAddPostHandler)
		wishlist.POST("/remove", controllers.WishListRemovePostHandler)
	}

	notify := g.Group("/notify")
	{
		notify.POST("/toggle", controllers.NotifyToggle)
	}

	categories := g.Group("/categories")
	{
		categories.GET("", controllers.SelectCategoriesPage)
		categories.GET("/:category", controllers.ListingsPage)

	}

	profile := g.Group("/profile")
	{
		profile.GET("", controllers.ProfileHandler)
	}
	payments := g.Group("/payments")
	{
		payments.POST("/webhook", controllers.Webhook)
	}

	decentro := g.Group("/payments/webhook/decentro")
	decentro.Use(middleware.DecentroWebhookAuth)
	{
		decentro.POST("/", controllers.DecentroWebhook)
	}

	pay := g.Group("/pay")
	{
		pay.GET("/:payID", controllers.RedirectToUPIDeeplinks)
	}

	ratings := g.Group("/reviews")
	{
		ratings.GET("/", controllers.GetReviewsForProduct) //Query Param Product ID
	}

	brand := g.Group("/brand")
	{
		brand.GET("/:brandID", controllers.BrandPage)
	}

	cart := g.Group("/cart")
	{
		cart.POST("/create/:variantID", controllers.CreateCart)
	}

	referral := g.Group("/referral")
	{
		referral.GET("", controllers.ReferralProgramPage)
		referral.GET("/join", controllers.JoinReferralProgramPage)
	}
}

func PrivateRoutes(g *gin.RouterGroup) {
	// All private routes
	address := g.Group("/address")
	{
		address.DELETE("", controllers.AddressDeleteHandler)
		address.GET("/:variantID", controllers.AddressGetHandler)
		address.POST("/:variantID", controllers.AddressPostHandler)
		address.GET("/pincodes/:pincode", controllers.GetDataByPincode)
	}

	addresses := g.Group("/addresses")
	{
		addresses.GET("", controllers.ManageAddresses)
		addresses.POST("", controllers.MakeNewAddress)
		addresses.PUT("", controllers.EditExistingAddress)
		addresses.POST("/default", controllers.UpdateDefaultAddress)
		addresses.GET("/new", controllers.NewAddressPage)
		addresses.GET("/edit/:addressID", controllers.EditAddressPage)
	}

	auth := g.Group("/auth")
	{
		auth.GET("/sign-out", controllers.SignOut)
	}

	cart := g.Group("/cart")
	{
		cart.GET("/summary", controllers.CartSummaryPage)
		cart.PUT("/:cartID/update/", controllers.UpdateItemQuantityInCart)
		// cart.POST("/wallet", controllers.ApplyWalletBalanceToCart)
		// cart.DELETE("/wallet", controllers.RemoveWalletBalanceFromCart)
	}

	order := g.Group("/order")
	{
		order.GET("", controllers.OrderList)
		order.GET("/:id", controllers.OrderSpecificPage)
		order.POST("/create", controllers.OrderPostHandlerV2)
		order.GET("/cancel/:orderID", controllers.OrderCancellationPage)
		order.POST("/cancel/:orderID", controllers.OrderCancellation)
		order.GET("/confirmed", controllers.OrderConfirmedHandler)
		order.GET("/success", controllers.OrderSuccessHandler)
		order.GET("/pending", controllers.OrderPendingHandler)
		order.GET("/failed", controllers.OrderFailureHandler)
		order.GET("/track", controllers.TrackOrder)
		// order.GET("/:productID", controllers.OrderGetHandler)
		// order.POST("/:productID/:addressID", controllers.OrderPostHandler)
		// order.POST("/:productID", controllers.OrderPostHandler)
		order.GET("/select-payment", controllers.SelectPaymentPage)
		order.POST("/select-payment", controllers.SelectPaymentPostHandler)
		order.GET("/coupon", controllers.CouponGetHandler)
		order.POST("/coupon/check", controllers.CheckCouponApplicability)
		order.DELETE("/coupon", controllers.RemoveAppliedCoupon)
	}

	payments := g.Group("/payments")
	{
		payments.GET("/redirect", controllers.PaymentRedirect)
		// a status endpoint which can be used to check the status of the payment
		// it'll be a websocket endpoint
		payments.GET("/status", controllers.PaymentStatus)
		payments.GET("/status/:paymentID", controllers.GetPaymentStatus)
		payments.GET("/save-card", controllers.SaveCardGetHandler)
		payments.POST("/save-card", controllers.SaveCardPostHandler)
	}

	ratings := g.Group("/review")
	{
		ratings.GET("/:productID", controllers.ReviewEntryPage)
		ratings.POST("", controllers.ReviewPostHandler2) //Product ID as query parameter, user from session
	}
	wishlist := g.Group("/wishlist")
	{
		wishlist.GET("", controllers.WishListGetHandler)
	}

	chat := g.Group("/chat")
	{
		chat.GET("", controllers.ChatHomePage)
		chat.GET("/:chatID", controllers.ChatSpecificPage)
		chat.GET("/create-or-get", controllers.CreateOrGetChat)
	}

	deals := g.Group("/deal")
	{
		deals.GET("/:variantID/:dealID", controllers.GetDealTeamSummary)
		deals.POST("", controllers.CreateNewTeamForDeal)
	}

	teams := g.Group("/team")
	{
		teams.POST("", controllers.JoinTeam) // teamID as query params
	}

	user := g.Group("/user")
	{
		user.POST("/join", controllers.JoinReferralProgram)
	}

	fakeAttempt := g.Group("/fake-attempt")
	{
		fakeAttempt.GET("", controllers.FakeAttemptController)
		fakeAttempt.GET("/download", controllers.DownloadFakeAttemptReportV2)
	}

	influencer := g.Group("/influencer")
	{
		influencer.GET("/authorize", controllers.InfluencerAuthorizePage)
		influencer.GET("/approval", controllers.InfluencerApprovalPage)
		influencer.GET("/instagram/oauth", controllers.GetInstagramOAuthURL)
		influencer.GET("/instagram/update", controllers.UpdateInstagramProfile)
		influencer.GET("/youtube/oauth", controllers.GetYouTubeOAuthURL)
		influencer.GET("/youtube/update", controllers.UpdateInstagramProfile)
		influencer.GET("/campaigns", controllers.AllInfluencerCampaigns)
		influencer.GET("/campaigns/my", controllers.MyInfluencerCampaigns)
		influencer.GET("/campaign/:id", controllers.InfluencerSpecificCampaign)
		influencer.GET("/campaign/apply/:id", controllers.CampaignApplicationPlatformsPage)
		influencer.GET("/campaign/apply/:id/2", controllers.CampaignApplicationSelectProductsPage)
		influencer.POST("/campaign/apply/:id/2", controllers.UpdateProductInInfluencerCampaignApplication)
		influencer.GET("/campaign/apply/:id/3", controllers.CampaignApplicationAddAddressPage)
		influencer.POST("/campaign/apply/:id/3", controllers.UpdateAddressInInfluencerApplication)
		influencer.GET("/campaign/apply/:id/4", controllers.CampaignApplicationTermsOfService)
		influencer.POST("/campaign/apply/:id/4", controllers.SubmitInfluencerCampaignApplication)
		influencer.GET("/campaign/apply/:id/submit", controllers.CampaignApplicationSuccessfulSubmission)
	}
}

func WebhooksRoutes(g *gin.RouterGroup) {
	g.POST("/tracking", shiprocket.TrackingEventReceived)
}
