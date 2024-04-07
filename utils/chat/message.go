package chat

import (
	"context"
	"fmt"

	stream_chat "github.com/GetStream/stream-chat-go/v5"
)

func SendMessage(chanType string, channel *stream_chat.Channel, userID string, message string) (string, error) {

	messageObject, err := channel.SendMessage(context.Background(), &stream_chat.Message{
		Text: message,
	}, userID)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return "", err
	}
	return messageObject.Message.ID, nil
}

func SendInitialSupportMessage(channel *stream_chat.Channel) (string, error) {
	message := "Hi there! How can we help you?"
	return SendMessage("support", channel, "roovo-support", message)
}

func SendInitialProductSupportMessage(channel *stream_chat.Channel, userID string, product_name string, brand_name string) (string, error) {
	message := fmt.Sprintf("Hey! I am interested in %s from %s. Can you help me out with some questions?", product_name, brand_name)
	return SendMessage("product-support", channel, userID, message)
}
