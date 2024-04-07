package chat

import (
	"context"
	"fmt"
	"time"

	stream "github.com/GetStream/stream-chat-go/v5"
	stream_chat "github.com/GetStream/stream-chat-go/v5"
)

func CreateUserToken(userID string) (string, error) {
	_, err := CreateUser(userID)
	if err != nil {
		return "", err
	}
	token, err := client.CreateToken(userID, time.Time{}, time.Now())
	if err != nil {
		return "", err
	}
	return token, nil
}

func RevokeUserToken(userID string) error {
	timeNow := time.Now()
	_, err := client.RevokeUserToken(context.Background(), userID, &timeNow)
	if err != nil {
		return err
	}
	return nil
}

func CreateUser(userID string) (*stream_chat.User, error) {
	userData := stream.User{
		ID: userID,
	}
	resp, err := client.UpsertUser(context.Background(), &userData)
	if err != nil {
		return nil, err
	}
	return resp.User, nil
}

func GetUser(userID string) (*stream_chat.User, error) {
	resp, err := client.QueryUsers(context.Background(), &stream_chat.QueryOption{
		UserID: userID,
		Filter: map[string]interface{}{
			"id": map[string]interface{}{
				"$eq": userID,
			},
		},
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Users) == 0 {
		return nil, nil
	}
	return resp.Users[0], nil
}

func InitRoovoSupportUser() error {
	existingUser, err := GetUser("roovo-support")
	fmt.Printf("existingUser: %v\n", existingUser)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return err
	}
	if existingUser == nil {
		_, err = CreateUser("roovo-support")
		if err != nil {
			return err
		}
	}
	return nil
}

func GetTokenForRoovoSupportUser() (string, error) {
	// err := RevokeUserToken("roovo-support")
	// if err != nil {
	// 	return "", err
	// }
	return CreateUserToken("roovo-support")
}
