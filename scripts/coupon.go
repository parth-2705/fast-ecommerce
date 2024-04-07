package scripts

import (
	"context"
	"fmt"
	"hermes/db"
	"hermes/models"

	"go.mongodb.org/mongo-driver/bson"
)

func createCoupon() (err error) {
	discount := models.CreateDiscount(69.0, 500.0, models.Percent)
	_, err = models.CreateCoupon(1, discount, "INFLUENZA69", models.ProductDiscount, []string{"ef0c813a-7065-408c-9d44-1847b74185e7"}, models.SingleUsePerUser)
	return
}

func MigrateOldCouponStructtoNew() (err error) {
	cursor, err := db.CouponCollection.Find(context.Background(), bson.M{})
	if err != nil {
		fmt.Printf("err1: %v\n", err)
		return
	}
	failed, passed := 0, 0

	var coupons []models.Coupon
	err = cursor.All(context.Background(), &coupons)
	if err != nil {
		fmt.Printf("err2: %v\n", err)
		return
	}
	for _, coupon := range coupons {
		coupon.ApplicableIDs = append(coupon.ApplicableIDs, coupon.ApplicableID)
		err = coupon.Update()
		if err != nil {
			failed++
			continue
		}
		passed++
	}
	fmt.Printf("failed: %v\n", failed)
	fmt.Printf("passed: %v\n", passed)
	fmt.Printf("total: %v\n", len(coupons))
	return
}
