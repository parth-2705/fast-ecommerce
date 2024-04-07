package shiprocket

import (
	"encoding/json"
	"fmt"
	"hermes/models"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tryamigo/themis"
)

type Courier struct {
	AirMaxWeight          string  `json:"air_max_weight"`
	BaseCourierID         any     `json:"base_courier_id"`
	BaseWeight            string  `json:"base_weight"`
	Blocked               int     `json:"blocked"`
	CallBeforeDelivery    string  `json:"call_before_delivery"`
	ChargeWeight          float64 `json:"charge_weight"`
	City                  string  `json:"city"`
	Cod                   int     `json:"cod"`
	CodCharges            int     `json:"cod_charges"`
	CodMultiplier         int     `json:"cod_multiplier"`
	Cost                  string  `json:"cost"`
	CourierCompanyID      int     `json:"courier_company_id"`
	CourierName           string  `json:"courier_name"`
	CourierType           string  `json:"courier_type"`
	CoverageCharges       int     `json:"coverage_charges"`
	CutoffTime            string  `json:"cutoff_time"`
	DeliveryBoyContact    string  `json:"delivery_boy_contact"`
	DeliveryPerformance   float64 `json:"delivery_performance"`
	Description           string  `json:"description"`
	Edd                   string  `json:"edd"`
	EntryTax              int     `json:"entry_tax"`
	EstimatedDeliveryDays string  `json:"estimated_delivery_days"`
	Etd                   string  `json:"etd"`
	EtdHours              int     `json:"etd_hours"`
	FreightCharge         float64 `json:"freight_charge"`
	ID                    int     `json:"id"`
	IsCustomRate          int     `json:"is_custom_rate"`
	IsHyperlocal          bool    `json:"is_hyperlocal"`
	IsInternational       int     `json:"is_international"`
	IsRtoAddressAvailable bool    `json:"is_rto_address_available"`
	IsSurface             bool    `json:"is_surface"`
	LocalRegion           int     `json:"local_region"`
	Metro                 int     `json:"metro"`
	MinWeight             float64 `json:"min_weight"`
	Mode                  int     `json:"mode"`
	Odablock              bool    `json:"odablock"`
	OtherCharges          int     `json:"other_charges"`
	Others                any     `json:"others"`
	PickupAvailability    int     `json:"pickup_availability"`
	PickupPerformance     float64 `json:"pickup_performance"`
	PickupPriority        string  `json:"pickup_priority"`
	PickupSupressHours    int     `json:"pickup_supress_hours"`
	PodAvailable          string  `json:"pod_available"`
	Postcode              string  `json:"postcode"`
	QcCourier             int     `json:"qc_courier"`
	Rank                  string  `json:"rank"`
	Rate                  float64 `json:"rate"`
	Rating                float64 `json:"rating"`
	RealtimeTracking      string  `json:"realtime_tracking"`
	Region                int     `json:"region"`
	RtoCharges            float64 `json:"rto_charges"`
	RtoPerformance        float64 `json:"rto_performance"`
	SecondsLeftForPickup  int     `json:"seconds_left_for_pickup"`
	State                 string  `json:"state"`
	SuppressDate          string  `json:"suppress_date"`
	SuppressText          string  `json:"suppress_text"`
	SuppressionDates      []any   `json:"suppression_dates"`
	SurfaceMaxWeight      string  `json:"surface_max_weight"`
	TrackingPerformance   float64 `json:"tracking_performance"`
	VolumetricMaxWeight   int     `json:"volumetric_max_weight"`
	WeightCases           float64 `json:"weight_cases"`
}
type ServiceabilityResponseBody struct {
	CompanyAutoShipmentInsuranceSetting bool `json:"company_auto_shipment_insurance_setting"`
	CovidZones                          struct {
		DeliveryZone any `json:"delivery_zone"`
		PickupZone   any `json:"pickup_zone"`
	} `json:"covid_zones"`
	Currency string `json:"currency"`
	Data     struct {
		AvailableCourierCompanies []Courier `json:"available_courier_companies"`
		ChildCourierID            any       `json:"child_courier_id"`
		IsRecommendationEnabled   int       `json:"is_recommendation_enabled"`
		RecommendedBy             struct {
			ID    int    `json:"id"`
			Title string `json:"title"`
		} `json:"recommended_by"`
		RecommendedCourierCompanyID    int `json:"recommended_courier_company_id"`
		ShiprocketRecommendedCourierID int `json:"shiprocket_recommended_courier_id"`
	} `json:"data"`
	DgCourier                    int   `json:"dg_courier"`
	EligibleForInsurance         int   `json:"eligible_for_insurance"`
	InsuraceOptedAtOrderCreation bool  `json:"insurace_opted_at_order_creation"`
	IsAllowTemplatizedPricing    bool  `json:"is_allow_templatized_pricing"`
	IsLatlong                    int   `json:"is_latlong"`
	LabelGenerateType            int   `json:"label_generate_type"`
	SellerAddress                []any `json:"seller_address"`
	Status                       int   `json:"status"`
	UserInsuranceManadatory      bool  `json:"user_insurance_manadatory"`
}

func CheckServiceability(deliveryPincode string, pickupPincode string, weight string) (deliveryTime string, err error) {
	token, err := GetShipRocketToken()
	if err != nil {
		return "", err
	}

	url := "https://apiv2.shiprocket.in/v1/external/courier/serviceability?pickup_postcode=" + pickupPincode + "&delivery_postcode=" + deliveryPincode + "&weight=" + weight + "&cod=1"
	headers := [][]string{{"Authorization", "Bearer " + token}}

	statusCode, responseBody, err := themis.HitAPIEndpoint2(url, http.MethodGet, []byte{}, headers, [][]string{})
	if err != nil {
		return "", err
	}
	if statusCode >= 400 {
		return "", fmt.Errorf("error: %s with status code: %d", string(responseBody), statusCode)
	}
	var respBody ServiceabilityResponseBody
	json.Unmarshal(responseBody, &respBody)
	if respBody.Status >= 400 {
		return "", fmt.Errorf("error: %s with status code: %d", string(responseBody), respBody.Status)
	}
	if len(respBody.Data.AvailableCourierCompanies) == 0 {
		return "", fmt.Errorf("could not find any available courier companies")
	}
	lowestDays := math.MaxInt
	lowestDaysString := strconv.Itoa(lowestDays)
	for _, val := range respBody.Data.AvailableCourierCompanies {
		estimatedDays, _ := strconv.Atoi(val.EstimatedDeliveryDays)
		if estimatedDays < lowestDays {
			lowestDays = estimatedDays
			lowestDaysString = strconv.Itoa(lowestDays)
		}
	}
	return lowestDaysString, nil
}

func GetCouriersForOrder(deliveryPincode string, pickupPincode string, weight string) (response []Courier, err error) {
	token, err := GetShipRocketToken()
	if err != nil {
		return []Courier{}, err
	}

	url := "https://apiv2.shiprocket.in/v1/external/courier/serviceability?pickup_postcode=" + pickupPincode + "&delivery_postcode=" + deliveryPincode + "&weight=" + weight + "&cod=1"
	headers := [][]string{{"Authorization", "Bearer " + token}}

	statusCode, responseBody, err := themis.HitAPIEndpoint2(url, http.MethodGet, []byte{}, headers, [][]string{})
	if err != nil {
		return []Courier{}, err
	}
	if statusCode >= 400 {
		return []Courier{}, fmt.Errorf("error: %s with status code: %d", string(responseBody), statusCode)
	}
	var respBody ServiceabilityResponseBody
	json.Unmarshal(responseBody, &respBody)
	if respBody.Status >= 400 {
		return []Courier{}, fmt.Errorf("error: %s with status code: %d", string(responseBody), respBody.Status)
	}
	if len(respBody.Data.AvailableCourierCompanies) == 0 {
		return []Courier{}, fmt.Errorf("could not find any available courier companies")
	}
	return respBody.Data.AvailableCourierCompanies, nil
}

func ServiceabilityController(c *gin.Context) {
	pickupPincode := c.Query("pickupPincode")
	deliveryPincode := c.Query("deliveryPincode")
	weight := c.Query("weight")
	time, err := CheckServiceability(deliveryPincode, pickupPincode, weight)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to check serviceability " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"delivery-time": time + " days"})
}

func GetAllCouriers(c *gin.Context) {
	sellerID := c.Query("seller")
	if sellerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Empty seller ID"})
		return
	}
	seller, err := models.GetSellerByID(sellerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid seller " + err.Error()})
	}
	deliveryPincode := c.Query("deliveryPincode")
	weight := c.Query("weight")
	couriers, err := GetCouriersForOrder(deliveryPincode, seller.PinCode, weight)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to get couriers " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"couriers": couriers})
}
