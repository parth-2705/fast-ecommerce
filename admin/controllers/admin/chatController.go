package controllers

import (
	"hermes/utils/chat"
	"hermes/utils/data"

	stream_chat "github.com/GetStream/stream-chat-go/v5"
	"github.com/gin-gonic/gin"
)

func AllChatsPage(c *gin.Context) {
	adminToken, err := chat.GetTokenForRoovoSupportUser()
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	channels, err := chat.GetChannelsForUser("roovo-support")
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	chatsData, channelTypes := chat.GetChatDataToRenderFromChannels(channels)

	var channel *stream_chat.Channel

	// get the channel id from the url
	channelID := c.Param("chatID")
	var userChatName string

	// get the channel object from the channel id
	if channelID != "home" {
		channel, err = chat.GetChannel(channelID)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		userChatName = data.MaskPhoneNumber(channel.CreatedBy.Name)
	} else {
		channel = &stream_chat.Channel{}
		userChatName = "Roovo Support"
	}
	c.HTML(200, "root", gin.H{
		"title":        "Roovo | Chats",
		"channels":     channels,
		"UserToken":    adminToken,
		"ChatWithName": userChatName,
		"UserID":       "roovo-support",
		"UserName":     "Roovo Support",
		"StreamKey":    chat.GetStreamAPIKey(),
		"chatsData":    chatsData,
		"channelTypes": channelTypes,
		"channel":      channel,
		"template":     "chats",
	})
}
