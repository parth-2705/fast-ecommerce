package controllers

import (
	"fmt"
	"hermes/models"
	"hermes/utils/chat"
	"net/http"

	stream_chat "github.com/GetStream/stream-chat-go/v5"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func ChatHomePage(c *gin.Context) {
	user, err := getUserObjectFromSession(c)
	if err != nil {
		fmt.Println(err)
		c.Redirect(http.StatusTemporaryRedirect, "/auth/sign-out")
		return
	}
	userProfile, err := user.GetProfile()
	if err != nil {
		fmt.Println("err getting user profile: ", err)
		c.AbortWithError(500, err)
		return
	}
	if userProfile.StreamUserToken == "" {
		userProfile.StreamUserToken, err = chat.CreateUserToken(user.ID)
		if err != nil {
			fmt.Println("err creating user token: ", err)
			c.AbortWithError(500, err)
			return
		}
		err = userProfile.Update()
	}

	chats, err := chat.GetChatsForUser(user.ID)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	all_channels, err := chat.GetChannelsForUser(user.ID)
	if err != nil {
		fmt.Println("err getting channels for user: ", err)
		c.AbortWithError(500, err)
		return
	}
	chatsData, channelTypes := chat.GetChatDataToRenderFromChannels(all_channels)
	c.HTML(http.StatusOK, "chat-home", gin.H{"chats": chats,
		"StreamKey":    chat.GetStreamAPIKey(),
		"all_channels": all_channels,
		"chatsData":    chatsData,
		"channelTypes": channelTypes,
		"user":         user,
		"UserToken":    userProfile.StreamUserToken,
		"UserID":       user.ID,
		"UserName":     user.Phone,
	})
}

func ChatSpecificPage(c *gin.Context) {
	user, err := getUserObjectFromSession(c)
	if err != nil {
		fmt.Println(err)
		c.Redirect(http.StatusTemporaryRedirect, "/auth/sign-out")
		// c.AbortWithError(401, err)
		return
	}

	userProfile, err := user.GetProfile()
	if err != nil {
		fmt.Println("err getting user profile: ", err)
		c.AbortWithError(500, err)
		return
	}
	if userProfile.StreamUserToken == "" {
		userProfile.StreamUserToken, err = chat.CreateUserToken(user.ID)
		if err != nil {
			fmt.Println("err creating user token: ", err)
			c.AbortWithError(500, err)
			return
		}
		err = userProfile.Update()
	}

	chatID := c.Param("chatID")

	fmt.Printf("chatID: %v\n", chatID)

	var chatObject *stream_chat.Channel

	if chatID == "support" {
		var isNew bool
		chatObject, isNew, err = chat.CreateOrGetChannelForUserSupport(user.ID)
		fmt.Println("Chat Object:", chatObject)
		if err != nil {
			fmt.Println("err creating new channel: ", err)
			c.AbortWithError(500, err)
			return
		}
		fmt.Printf("isNew: %v\n", isNew)
		fmt.Printf("err: %v\n", err)
		if isNew {
			fmt.Println("Sending Initial Support Message")
			_, err = chat.SendInitialSupportMessage(chatObject)
			if err != nil {
				fmt.Println("err sending initial support message: ", err)
				c.AbortWithError(500, err)
				return
			}
		}
	}

	if chatID == "product-support" {
		// get variant id from query
		variantID := c.Query("variant_id")
		if variantID == "" {
			c.AbortWithError(500, mongo.ErrNoDocuments)
			return
		}
		variant, err := models.GetVariant(variantID)
		if err != nil {
			fmt.Println("err getting variant: ", err)
			c.AbortWithError(500, err)
			return
		}
		product, err := models.GetProduct(variant.ProductID)
		if err != nil {
			fmt.Println("err getting product: ", err)
			c.AbortWithError(500, err)
			return
		}
		var isNew bool
		chatObject, isNew, err = chat.CreateOrGetChannelForProductSupport(user.ID, variantID)
		if err != nil {
			fmt.Println("err creating new channel: ", err)
			c.AbortWithError(500, err)
			return
		}
		fmt.Printf("isNew: %v\n", isNew)
		fmt.Printf("err: %v\n", err)
		if isNew {
			fmt.Println("Sending Initial Support Message")
			_, err = chat.SendInitialProductSupportMessage(chatObject, user.ID, product.Name, product.Brand.Name)
			if err != nil {
				fmt.Println("err sending initial support message: ", err)
				c.AbortWithError(500, err)
				return
			}
		}
	}

	// chatChannel, err := chat.GetChatByID(chatID)

	// chats := chat.GetMessagesForChatID(chatID, user.ID)
	c.HTML(http.StatusOK, "chat-specific", gin.H{"channel": chatObject, "UserToken": userProfile.StreamUserToken, "UserID": user.ID, "UserName": user.Phone, "StreamKey": chat.GetStreamAPIKey(), "ChatID": chatID})
}

func CreateOrGetChat(c *gin.Context) {
	user, err := getUserObjectFromSession(c)
	if err != nil {
		fmt.Println(err)
		c.Redirect(http.StatusTemporaryRedirect, "/auth/sign-out")
		// c.AbortWithError(401, err)
		return
	}
	userProfile, err := user.GetProfile()
	if err != nil {
		fmt.Println("err getting user profile: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"response": "Error in getting user profile " + err.Error()})
		return
	}
	if userProfile.StreamUserToken == "" {
		userProfile.StreamUserToken, err = chat.CreateUserToken(user.ID)
		if err != nil {
			fmt.Println("err creating user token: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"response": "Error in creating stream token " + err.Error()})
			return
		}
		err = userProfile.Update()
	}

	chat.UpsertUserInStream(user)
	var chatChannel chat.ChatChannel

	chatType := c.Query("type")
	id := ""
	// if chatType == "support" {
	// 	id = "roovo-support"
	// } else {
	// 	id = c.Query("id")
	// }
	id = "roovo-support"

	channelID, doesChatExist, err := chat.CheckIfChatExists(user.ID, id, chatType)
	if err != nil && err != mongo.ErrNoDocuments {
		c.JSON(http.StatusServiceUnavailable, gin.H{"response": "Error in checking if chat exists " + err.Error()})
		return
	}
	if !doesChatExist {
		chatChannel = chat.CreateChannel(user.ID, id, chatType)

		err = chatChannel.SaveToDB()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"response": "Error saving chat to DB " + err.Error()})
			return
		}
		channelID = chatChannel.ID
	}
	c.JSON(http.StatusOK, gin.H{"channelID": channelID})
	return
}
