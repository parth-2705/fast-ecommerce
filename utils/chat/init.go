package chat

import (
	"context"
	"hermes/db"
	"hermes/models"
	"hermes/utils/data"
	"os"

	stream "github.com/GetStream/stream-chat-go/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var client *stream.Client

type ChatChannel struct {
	ID          string `json:"_id" bson:"_id"`
	Type        string `json:"type" bson:"type"`
	UserID      string `json:"userID" bson:"userID"`
	ResponderID string `json:"responderID" bson:"responderID"`
}

// type Chat struct {
// 	ID          string `json:"_id" bson:"_id"`
// 	Type        string `json:"type" bson:"type"`
// 	UserID      string `json:"userID" bson:"userID"`
// 	ResponderID string `json:"responderID" bson:"responderID"`
// }

func Init() {
	client, _ = stream.NewClient(os.Getenv("STREAM_API_KEY"), os.Getenv("STREAM_API_SECRET"))
}

func GetStreamAPIKey() string {
	return os.Getenv("STREAM_API_KEY")
}

func UpsertUserInStream(user models.User) {
	client.UpsertUser(context.Background(), &stream.User{ID: user.ID, Name: user.Name})
	return
}

func CreateChannel(userID string, responderID string, chatType string) (chatChannel ChatChannel) {

	var channelType string
	channelType = chatType
	chatChannel.ID = data.GetUUIDString("channel")
	chatChannel.UserID = userID
	chatChannel.ResponderID = responderID
	chatChannel.Type = channelType

	chatChannel.CreateChannel()
	return
}

func (channel ChatChannel) CreateChannel() {
	client.CreateChannel(context.Background(), channel.Type, channel.ID, channel.UserID, &stream.ChannelRequest{
		Members: []string{channel.UserID, channel.ResponderID},
	})
	return
}

func (channel ChatChannel) SaveToDB() (err error) {
	_, err = db.ChatChannelCollection.InsertOne(context.Background(), channel)
	return
}

func GetChatsForUser(userID string) (chats []ChatChannel, err error) {
	cur, err := db.ChatChannelCollection.Find(context.Background(), bson.M{"userID": userID})
	if err != nil && err != mongo.ErrNoDocuments {
		return
	}
	err = cur.All(context.Background(), &chats)
	return
}

func GetChatByID(channelID string) (chat ChatChannel, err error) {
	err = db.ChatChannelCollection.FindOne(context.Background(), bson.M{"_id": channelID}).Decode(&chat)
	return
}

func CheckIfChatExists(userID string, responderID string, chatType string) (channelID string, doesChatExist bool, err error) {
	var chatChannel ChatChannel
	err = db.ChatChannelCollection.FindOne(context.Background(), bson.M{"userID": userID, "responderID": responderID, "type": chatType}).Decode(&chatChannel)
	if err != nil && err != mongo.ErrNoDocuments {
		return
	}
	channelID = chatChannel.ID
	doesChatExist = (chatChannel.ID != "")
	return
}

func GetMessagesForChatID(channelID string, userID string) (chats string) {
	return
}
