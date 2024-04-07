package chat

import (
	"context"
	"fmt"
	"hermes/models"
	"hermes/utils/data"

	stream_chat "github.com/GetStream/stream-chat-go/v5"
)

type ChatData struct {
	ChatID          string
	ChatType        string
	LastMessage     string
	LastMessageTime string
	MobileNumber    string
	Variant         *models.Variation
	Product         *models.Product
}

func CreateNewChannel(channelID string, channelType string, userId string, members []string, extraData map[string]interface{}) (*stream_chat.Channel, error) {
	response, err := client.CreateChannel(context.Background(), channelType, channelID, userId, &stream_chat.ChannelRequest{
		Members:   members,
		ExtraData: extraData,
	})
	if err != nil {
		return nil, err
	}

	if response.Channel == nil {
		fmt.Println("response.Channel is nil")
	}

	return response.Channel, nil
}

// func to generate unique channel id
func generateChannelID(chanType string) string {
	switch chanType {
	case "support":
		// return "roovo-support" + 6 random digits
		random_string, err := data.GenerateRandomHex(6)
		if err != nil {
			return ""
		}
		return "roovo-support-" + random_string
	default:
		return ""
	}

}

func CreateOrGetChannelForUserSupport(userID string) (*stream_chat.Channel, bool, error) {
	fmt.Printf("userID: %v\n", userID)
	resp, err := client.QueryChannels(context.Background(), &stream_chat.QueryOption{
		Filter: map[string]interface{}{
			"$and": []map[string]interface{}{
				{
					"members": map[string]interface{}{
						"$in": []string{userID},
					},
				},
				{
					"members": map[string]interface{}{
						"$in": []string{"roovo-support"},
					},
				},
				{
					"type": "support",
				},
			},
		},
		Sort:  []*stream_chat.SortOption{{Field: "last_message_at", Direction: 1}}, // sorting direction (1 or -1)
		Limit: 1,
	})
	if err != nil {
		fmt.Println(err, "err")
		return nil, false, err
	}
	fmt.Println(resp.Channels, "resp.Channels")
	if len(resp.Channels) > 0 {
		return resp.Channels[0], false, err
	} else {
		fmt.Println("creating new channel")
		channel, err := CreateNewChannelForUserSupport(userID)
		return channel, true, err
	}
}

func CreateOrGetChannelForProductSupport(userID string, variantID string) (*stream_chat.Channel, bool, error) {
	resp, err := client.QueryChannels(context.Background(), &stream_chat.QueryOption{
		Filter: map[string]interface{}{
			"cid": map[string]interface{}{
				"$eq": "product-support:" + data.Md5Hash(userID+"-"+variantID),
			},
		},
		Sort:  []*stream_chat.SortOption{{Field: "last_message_at", Direction: 1}}, // sorting direction (1 or -1)
		Limit: 1,
	})
	if err != nil {
		fmt.Println(err, "err")
		return nil, false, err
	}
	if len(resp.Channels) > 0 {
		return resp.Channels[0], false, err
	} else {
		channel, err := CreateNewChannelForProductSupport(userID, variantID)
		return channel, true, err
	}
}

func CreateNewChannelForUserSupport(userID string) (*stream_chat.Channel, error) {
	channelID := "roovo-" + userID
	if channelID == "" {
		return nil, fmt.Errorf("error generating channel id")
	}
	channelType := "support"
	members := []string{userID, "roovo-support"}
	return CreateNewChannel(channelID, channelType, userID, members, nil)
}

func CreateNewChannelForProductSupport(userID string, variantID string) (*stream_chat.Channel, error) {
	channelID := userID + "-" + variantID
	// hashing the channelID to make it unique and not contain any sensitive data
	channelID = data.Md5Hash(channelID)
	if channelID == "" {
		return nil, fmt.Errorf("error generating channel id")
	}
	channelType := "product-support"
	members := []string{userID, "roovo-support"}
	extraData := map[string]interface{}{
		"variant_id": variantID,
	}
	return CreateNewChannel(channelID, channelType, userID, members, extraData)
}

func GetChannelsForUser(userID string) ([]*stream_chat.Channel, error) {
	response, err := client.QueryChannels(context.Background(), &stream_chat.QueryOption{
		Filter: map[string]interface{}{
			"members": map[string]interface{}{
				"$in": []string{userID},
			},
		},
		Sort: []*stream_chat.SortOption{{Field: "last_message_at", Direction: 1}}, // sorting direction (1 or -1)
	})
	if err != nil {
		return nil, err
	}
	return response.Channels, nil
}

func GetChatDataToRenderFromChannels(channels []*stream_chat.Channel) ([]ChatData, []string) {
	var chatsData []ChatData
	var channelTypes []string

	for _, channel := range channels {
		messages := channel.Messages
		fmt.Printf("channel: %v\n", channel)
		fmt.Printf("channel.ExtraData: %v\n", channel.ExtraData)

		if len(messages) > 0 {
			lastMessage := messages[len(messages)-1]
			var variant *models.Variation
			channelType := channel.Type
			fmt.Printf("channelType: %v\n", channelType)
			variantID, ok := channel.ExtraData["variant_id"]
			if ok {
				fmt.Printf("variantID: %v\n", variantID)
				variation, err := models.GetVariant(variantID.(string))
				if err != nil {
					fmt.Printf("err: %+v\n", err)
				}
				fmt.Printf("variant: %+v\n", variant)
				if variation.ID != "" {
					variant = &variation
				}
			}
			var lastMessageDisplayText string
			if lastMessage.Text != "" {
				lastMessageDisplayText = lastMessage.Text
			} else if lastMessage.Attachments != nil {
				lastMessageDisplayText = "Attachment"
			} else {
				lastMessageDisplayText = "Message"
			}
			chatData := ChatData{
				ChatID:          channel.ID,
				ChatType:        channel.Type,
				LastMessage:     lastMessageDisplayText,
				LastMessageTime: lastMessage.CreatedAt.Format("03:04 PM"),
				MobileNumber:    data.MaskPhoneNumber(channel.CreatedBy.Name),
			}
			if variant != nil {
				fmt.Println("variant: ", variant)
				chatData.Variant = variant

				product, err := models.GetProduct(chatData.Variant.ProductID)
				if err != nil {
					fmt.Printf("err: %+v\n", err)
				}
				chatData.Product = &product
			}
			if variant == nil && channelType != "support" {
				continue
			}

			chatsData = append(chatsData, chatData)
			if !data.Contains(channelTypes, channel.Type) {
				channelTypes = append(channelTypes, channel.Type)
			}
		}
	}
	return chatsData, channelTypes
}

func GetChannel(channelID string) (*stream_chat.Channel, error) {
	response, err := client.QueryChannels(context.Background(), &stream_chat.QueryOption{
		Filter: map[string]interface{}{
			"id": channelID,
		},
	})
	if err != nil {
		return nil, err
	}
	if len(response.Channels) > 0 {
		return response.Channels[0], nil
	}
	return nil, fmt.Errorf("channel not found")
}
