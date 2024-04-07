package controllers

import (
	"hermes/utils/amplitude"
	"hermes/utils/network"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Link struct {
	Url      string `json:"url" bson:"url"`
	Text     string `json:"text" bson:"text"`
	Icon     string `json:"icon" bson:"icon"`
	IsPublic bool   `json:"isPublic" bson:"isPublic"`
}

func ProfileHandler(c *gin.Context) {
	go amplitude.TrackEventByAuth("Profile Page", c)

	user, err := getUserObjectFromSession(c)
	loggedIn := true
	if err != nil {
		loggedIn = false
	}

	if network.MobileRequest(c) {
		c.JSON(http.StatusOK, gin.H{
			"phone": user.Phone,
		})

		return
	}

	var referralPageURL string
	var referralPageText string

	if user.HasJoinedReferralProgram {
		referralPageURL = "/referral"
		referralPageText = "Ambassador program"
	} else {
		referralPageURL = "/referral/join"
		referralPageText = "Ambassador Program"
	}

	c.HTML(http.StatusOK, "profile", gin.H{
		"profile": "",
		"profileLinks": []Link{{
			Url:  "/order",
			Text: "Orders",
			Icon: "/static/assets/OrderListIcon.svg",
		}, {
			Url:  "/addresses",
			Text: "Manage Addresses",
			Icon: "/static/assets/manageAddressIcon.svg",
		}, {
			Url:  "/wishlist",
			Text: "Wishlist",
			Icon: "/static/assets/wishlistIcon.svg",
		}, {
			Url:      "/contact-us",
			Text:     "Contact Us",
			Icon:     "/static/assets/contactUs.svg",
			IsPublic: true,
		},
			{
				Url:      referralPageURL,
				Text:     referralPageText,
				Icon:     "/static/assets/referralProgram.svg",
				IsPublic: true,
			},
			{
				Url:      "/influencer",
				Text:     "Influencer Program",
				Icon:     "/static/icons/InfluencerIcon.png",
				IsPublic: true,
			},
		},
		"publicLinks": []Link{{
			Url:  "/about-us",
			Text: "About Us",
		}, {
			Url:  "/privacy-policy",
			Text: "Privacy Policy",
		}, {
			Url:  "/terms-and-conditions",
			Text: "Terms & Conditions",
		}, {
			Url:  "/online-registration-policy",
			Text: "User Policies",
		},
			{
				Url:  "/return-policy",
				Text: "Return Policy",
			}},
		"phone":    user.Phone,
		"loggedIn": loggedIn,
	})
}
