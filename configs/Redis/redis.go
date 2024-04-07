package Redis

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func Connect() (err error) {
	Client = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	status := Client.Ping(context.Background())
	err = status.Err()
	if err != nil {
		return
	}
	return
}

func DeleteProductCacheByID(productID string) (err error) {
	res := Client.Del(context.Background(), productID)
	err = res.Err()
	return
}

func DeleteKeyFromRedis(key string) (err error) {
	res := Client.Del(context.Background(), key)
	err = res.Err()
	return
}

func AllPincodesOfBrand(key string) []string {
	res := Client.SMembers(context.Background(), key)
	return res.Val()
}

func SetPincodesToBrand(brandID string, pincodes []string) (err error) {
	bytes, _ := json.Marshal(pincodes)
	pincodesInterface := []interface{}{}
	err = json.Unmarshal(bytes, &pincodesInterface)
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}
	err = Client.SAdd(context.Background(), brandID, pincodesInterface...).Err()
	return
}

func CheckIfPincodesExistForBrand(brandID string, pincode string) (exists bool, err error) {
	resSize := Client.SCard(context.Background(), brandID)
	err = resSize.Err()
	if err != nil {
		return
	}
	if resSize.Val() == 0 {
		exists = true
		return
	}
	res := Client.SIsMember(context.Background(), brandID, pincode)
	err = res.Err()
	if err != nil {
		return
	}
	exists = res.Val()
	return
}
