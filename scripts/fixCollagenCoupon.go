package scripts

import "hermes/models"

func fixCollagenCoupon(newProductID string, oldProductID string) (err error) {

	coupons, err := models.GetAllCoupons()
	if err != nil {
		return err
	}

	for _, coupon := range coupons {

		if coupon.ApplicableID == oldProductID {
			coupon.ApplicableIDs = append(coupon.ApplicableIDs, newProductID)
			err = coupon.Update()
			if err != nil {
				continue
			}
		}

	}

	return nil
}
