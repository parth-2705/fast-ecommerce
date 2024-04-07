package models

import (
	"context"
	"fmt"
	"hermes/db"
	"hermes/utils/data"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ShippingCharges struct {
	ID                string `json:"id" bson:"_id"`
	OrderID           string `json:"orderID" bson:"orderID"`
	ShiprocketOrderID string `json:"shiprocketOrderID" bson:"shiprocketOrderID"`
	Data              struct {
		AvailableCourierCompanies []ShiprocketCourier `json:"available_courier_companies,omitempty"`
		ChildCourierID            any                 `json:"child_courier_id,omitempty"`
		IsRecommendationEnabled   any                 `json:"is_recommendation_enabled,omitempty"`
		RecommendedBy             struct {
			ID    any `json:"id,omitempty"`
			Title any `json:"title,omitempty"`
		} `json:"recommended_by,omitempty"`
		RecommendedCourierCompanyID    any `json:"recommended_courier_company_id,omitempty"`
		ShiprocketRecommendedCourierID any `json:"shiprocket_recommended_courier_id,omitempty"`
	} `json:"data,omitempty"`
	EligibleForInsurance any `json:"eligible_for_insurance,omitempty"`
	Status               int `json:"status,omitempty"`
}

type ShiprocketCourier struct {
	AirMaxWeight          any `json:"air_max_weight,omitempty"`
	BaseCourierID         any `json:"base_courier_id,omitempty"`
	BaseWeight            any `json:"base_weight,omitempty"`
	Blocked               any `json:"blocked,omitempty"`
	CallBeforeDelivery    any `json:"call_before_delivery,omitempty"`
	ChargeWeight          any `json:"charge_weight,omitempty"`
	City                  any `json:"city,omitempty"`
	Cod                   any `json:"cod,omitempty"`
	CodCharges            any `json:"cod_charges,omitempty"`
	CodMultiplier         any `json:"cod_multiplier,omitempty"`
	Cost                  any `json:"cost,omitempty"`
	CourierCompanyID      any `json:"courier_company_id,omitempty"`
	CourierName           any `json:"courier_name,omitempty"`
	CourierType           any `json:"courier_type,omitempty"`
	CoverageCharges       any `json:"coverage_charges,omitempty"`
	CutoffTime            any `json:"cutoff_time,omitempty"`
	DeliveryBoyContact    any `json:"delivery_boy_contact,omitempty"`
	DeliveryPerformance   any `json:"delivery_performance,omitempty"`
	Description           any `json:"description,omitempty"`
	Edd                   any `json:"edd,omitempty"`
	EntryTax              any `json:"entry_tax,omitempty"`
	EstimatedDeliveryDays any `json:"estimated_delivery_days,omitempty"`
	Etd                   any `json:"etd,omitempty"`
	EtdHours              any `json:"etd_hours,omitempty"`
	FreightCharge         any `json:"freight_charge,omitempty"`
	ID                    any `json:"id,omitempty"`
	IsCustomRate          any `json:"is_custom_rate,omitempty"`
	IsHyperlocal          any `json:"is_hyperlocal,omitempty"`
	IsInternational       any `json:"is_international,omitempty"`
	IsRtoAddressAvailable any `json:"is_rto_address_available,omitempty"`
	IsSurface             any `json:"is_surface,omitempty"`
	LocalRegion           any `json:"local_region,omitempty"`
	Metro                 any `json:"metro,omitempty"`
	MinWeight             any `json:"min_weight,omitempty"`
	Mode                  any `json:"mode,omitempty"`
	Odablock              any `json:"odablock,omitempty"`
	OtherCharges          any `json:"other_charges,omitempty"`
	Others                any `json:"others,omitempty"`
	PickupAvailability    any `json:"pickup_availability,omitempty"`
	PickupPerformance     any `json:"pickup_performance,omitempty"`
	PickupPriority        any `json:"pickup_priority,omitempty"`
	PickupSupressHours    any `json:"pickup_supress_hours,omitempty"`
	PodAvailable          any `json:"pod_available,omitempty"`
	Postcode              any `json:"postcode,omitempty"`
	QcCourier             any `json:"qc_courier,omitempty"`
	Rank                  any `json:"rank,omitempty"`
	Rate                  any `json:"rate,omitempty"`
	Rating                any `json:"rating,omitempty"`
	RealtimeTracking      any `json:"realtime_tracking,omitempty"`
	Region                any `json:"region,omitempty"`
	RtoCharges            any `json:"rto_charges,omitempty"`
	RtoPerformance        any `json:"rto_performance,omitempty"`
	SecondsLeftForPickup  any `json:"seconds_left_for_pickup,omitempty"`
	State                 any `json:"state,omitempty"`
	SuppressDate          any `json:"suppress_date,omitempty"`
	SuppressText          any `json:"suppress_text,omitempty"`
	SurfaceMaxWeight      any `json:"surface_max_weight,omitempty"`
	TrackingPerformance   any `json:"tracking_performance,omitempty"`
	VolumetricMaxWeight   any `json:"volumetric_max_weight,omitempty"`
	WeightCases           any `json:"weight_cases,omitempty"`
}

func (shippingCharges ShippingCharges) CreateIndexes() error {
	mongoIndexes := []mongo.IndexModel{}

	orderIDModel := mongo.IndexModel{
		Keys: bson.M{
			"orderID": 1, // 1 for ascending, -1 for descending
		},
	}
	mongoIndexes = append(mongoIndexes, orderIDModel)

	shiprocketOrderIDModel := mongo.IndexModel{
		Keys: bson.M{
			"shiprocketOrderID": 1, // 1 for ascending, -1 for descending
		},
	}
	mongoIndexes = append(mongoIndexes, shiprocketOrderIDModel)

	_, err := db.ShippingCollection.Indexes().CreateMany(context.Background(), mongoIndexes)
	if err != nil {
		// handle error
		if strings.Contains(err.Error(), "Index with name") && strings.Contains(err.Error(), "already exists") {
			fmt.Println("Indexes already exist")
		} else {
			// Handle other errors
			fmt.Println("Error creating indexes:", err)
			return err
		}
	}
	return nil
}

func (shippingCharges *ShippingCharges) Create() error {

	shippingCharges.ID = data.GetUUIDString("charges")
	if len(shippingCharges.OrderID) == 0 {
		return fmt.Errorf("order id empty")
	}

	if len(shippingCharges.ShiprocketOrderID) == 0 {
		return fmt.Errorf("shiprocket order id empty")
	}

	_, err := db.ShippingCharges.InsertOne(context.Background(), shippingCharges)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func GetShippingChargesByOrderID(shiprocketOrderID string) (ShippingCharges, error) {

	var shippingCharges ShippingCharges
	if len(shiprocketOrderID) == 0 {
		return shippingCharges, fmt.Errorf("order id empty")
	}

	err := db.ShippingCharges.FindOne(context.Background(), bson.M{"shiprocketOrderID": shiprocketOrderID}).Decode(&shippingCharges)
	if err != nil {
		return shippingCharges, err
	}

	return shippingCharges, nil
}
