package models

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"hermes/configs/Mysql"
	"hermes/db"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"gorm.io/gorm/clause"
)

func GormScan(data reflect.Value, value interface{}) error {
	if value == nil {
		return nil
	}

	var byteSlice []byte
	switch v := value.(type) {
	case []byte:
		if len(v) > 0 {
			byteSlice = make([]byte, len(v))
			copy(byteSlice, v)
		}
	case string:
		byteSlice = []byte(v)
	default:
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	err := json.Unmarshal(byteSlice, data.Interface())
	return err
}

type FullShiprocketOrder struct {
	ID              string `json:"_id" bson:"_id"`
	SRID            int    `json:"id" bson:"srid"`
	ChannelID       int    `json:"channel_id"`
	ChannelName     string `json:"channel_name"`
	BaseChannelCode string `json:"base_channel_code"`
	IsInternational int    `json:"is_international"`
	IsDocument      int    `json:"is_document"`
	ChannelOrderID  string `json:"channel_order_id"`
	CustomerName    string `json:"customer_name"`
	CustomerEmail   string `json:"customer_email"`
	CustomerPhone   string `json:"customer_phone"`
	CustomerAddress string `json:"customer_address"`
	CustomerCity    string `json:"customer_city"`
	CustomerState   string `json:"customer_state"`
	CustomerPincode string `json:"customer_pincode"`
	CustomerCountry string `json:"customer_country"`
	PickupCode      string `json:"pickup_code"`
	// PickupLocation            string             `json:"pickup_location"`
	// PickupLocationID          int32              `json:"pickup_location_id"`
	// PickupID                  string             `json:"pickup_id"`
	// ShipType                  string             `json:"ship_type"`
	// CourierMode string `json:"courier_mode"`
	// Currency                  string             `json:"currency"`
	// CountryCode               int                `json:"country_code"`
	// ExchangeRateUsd           int                `json:"exchange_rate_usd"`
	// ExchangeRateInr           int                `json:"exchange_rate_inr"`
	// StateCode                 int                `json:"state_code"`
	PaymentStatus  string `json:"payment_status"`
	DeliveryCode   string `json:"delivery_code"`
	Total          int    `json:"total"`
	TotalInr       int    `json:"total_inr"`
	TotalUsd       int    `json:"total_usd"`
	NetTotal       string `json:"net_total"`
	OtherCharges   string `json:"other_charges"`
	OtherDiscounts string `json:"other_discounts"`
	// GiftwrapCharges           string             `json:"giftwrap_charges"`
	Expedited int    `json:"expedited"`
	SLA       string `json:"sla"`
	Cod       int    `json:"cod"`
	Tax       int    `json:"tax"`
	// TotalKeralaCess           string             `json:"total_kerala_cess"`
	Discount      int    `json:"discount"`
	Status        string `json:"status"`
	StatusCode    int    `json:"status_code"`
	MasterStatus  string `json:"master_status"`
	PaymentMethod string `json:"payment_method"`
	// PurposeOfShipment         int                `json:"purpose_of_shipment"`
	ChannelCreatedAt string         `json:"channel_created_at"`
	CreatedAt        string         `json:"created_at"`
	OrderDate        string         `json:"order_date"`
	UpdatedAt        string         `json:"updated_at"`
	Products         SRProductSlice `json:"products"`
	InvoiceNo        string         `json:"invoice_no"`
	Shipments        SRShipment     `json:"shipments"`
	AwbData          SRAWBData      `json:"awb_data"`
	// ReturnPickupData SRReturnPickupData `json:"return_pickup_data"`
	AllowReturn     int    `json:"allow_return"`
	IsReturn        int    `json:"is_return"`
	IsIncomplete    int    `json:"is_incomplete"`
	CouponIsVisible bool   `json:"coupon_is_visible"`
	Coupons         string `json:"coupons"`
	// BillingCity               string             `json:"billing_city"`
	// BillingName               string             `json:"billing_name"`
	// BillingEmail              string             `json:"billing_email"`
	// BillingPhone              string             `json:"billing_phone"`
	// BillingAlternatePhone     string             `json:"billing_alternate_phone"`
	// BillingStateName          string             `json:"billing_state_name"`
	// BillingAddress            string             `json:"billing_address"`
	// BillingCountryName        string             `json:"billing_country_name"`
	// BillingPincode            string             `json:"billing_pincode"`
	// BillingAddress2           string             `json:"billing_address_2"`
	// BillingMobileCountryCode  string             `json:"billing_mobile_country_code"`
	// IsdCode                   string             `json:"isd_code"`
	// BillingStateID            string             `json:"billing_state_id"`
	// BillingCountryID          string             `json:"billing_country_id"`
	// FreightDescription        string             `json:"freight_description"`
	// ResellerName              string             `json:"reseller_name"`
	// ShippingIsBilling         int                `json:"shipping_is_billing"`
	APIOrderID      string       `json:"api_order_id"`
	AllowMultiship  int          `json:"allow_multiship"`
	Others          SROthersData `json:"others"`
	IsOrderVerified int          `json:"is_order_verified"`
	// ExtraInfo                 SRExtraInfo        `json:"extra_info"`
	// Dup                       int                `json:"dup"`
	// IsBlackboxSeller          bool               `json:"is_blackbox_seller"`
	// ShippingMethod            string             `json:"shipping_method"`
	// RefundDetail              SRRefundDetail     `json:"refund_detail"`
	EwayBillNumber string `json:"eway_bill_number"`
	// EwayBillURL               string             `json:"eway_bill_url"`
	// EwayRequired              bool               `json:"eway_required"`
	// SellerCanEdit             bool               `json:"seller_can_edit"`
	// SellerCanCancell          bool               `json:"seller_can_cancell"`
	// IsPostShipStatus          bool               `json:"is_post_ship_status"`
	// OrderTag                  string             `json:"order_tag"`
	// QcStatus                  string             `json:"qc_status"`
	// QcReason                  string             `json:"qc_reason"`
	// QcImage                   string             `json:"qc_image"`
	// ChangePaymentMode         bool               `json:"change_payment_mode"`
	EtdDate                   string    `json:"etd_date"`
	OutForDeliveryDate        string    `json:"out_for_delivery_date"`
	DeliveredDate             string    `json:"delivered_date"`
	DeliveredDateTimeStamp    time.Time `json:"delivered_date_timestamp"`
	RTODeliveredDateTimeStamp time.Time `json:"rto_delivered_date_timestamp"`
	RemittanceDate            string    `json:"remittance_date"`
	RemittanceUtr             string    `json:"remittance_utr"`
	RemittanceStatus          string    `json:"remittance_status"`
	// InsuranceExcluded         bool               `json:"insurance_excluded"`
	// CanEditDimension          bool               `json:"can_edit_dimension"`
	// CustomerAddress2  any    `json:"customer_address_2"`
	// SubStatus         any    `json:"sub_status"`
	// CompanyLogo       any                `json:"company_logo"`
	// Errors                   any    `json:"errors"`
	// PaymentCode              any    `json:"payment_code"`
	// OtherSubOrders           []any              `json:"other_sub_orders"`
	// PickupAddress            []any              `json:"pickup_address"`
	// SellerRequest            any                `json:"seller_request"`
}

func (FullShiprocketOrder) TableName() string {
	return "shiprocket_orders"
}

type SRProducts struct {
	ID                     int     `json:"id"`
	OrderID                int     `json:"order_id"`
	ProductID              int     `json:"product_id"`
	Name                   string  `json:"name"`
	Sku                    string  `json:"sku"`
	Description            string  `json:"description"`
	ChannelOrderProductID  string  `json:"channel_order_product_id"`
	ChannelSku             string  `json:"channel_sku"`
	Hsn                    string  `json:"hsn"`
	Model                  any     `json:"model"`
	Manufacturer           any     `json:"manufacturer"`
	Brand                  string  `json:"brand"`
	Color                  string  `json:"color"`
	Size                   any     `json:"size"`
	CustomField            string  `json:"custom_field"`
	CustomFieldValue       string  `json:"custom_field_value"`
	CustomFieldValueString string  `json:"custom_field_value_string"`
	Weight                 int     `json:"weight"`
	Dimensions             string  `json:"dimensions"`
	Price                  float32 `json:"price"`
	Cost                   float32 `json:"cost"`
	Mrp                    float32 `json:"mrp"`
	Quantity               int     `json:"quantity"`
	ReturnableQuantity     int     `json:"returnable_quantity"`
	Tax                    float32 `json:"tax"`
	Status                 int     `json:"status"`
	NetTotal               float32 `json:"net_total"`
	Discount               float32 `json:"discount"`
	SellingPrice           float32 `json:"selling_price"`
	TaxPercentage          float32 `json:"tax_percentage"`
	DiscountIncludingTax   int     `json:"discount_including_tax"`
	ChannelCategory        string  `json:"channel_category"`
	PackagingMaterial      string  `json:"packaging_material"`
	AdditionalMaterial     string  `json:"additional_material"`
	IsFreeProduct          string  `json:"is_free_product"`
}

func (SRProducts) GormDataType() string {
	return "json"
}

func (data SRProducts) Value() (driver.Value, error) {
	jsonifiedData, _ := json.Marshal(data)
	return string(jsonifiedData), nil
}

func (data *SRProducts) Scan(value interface{}) error {
	err := GormScan(reflect.ValueOf(data), value)
	if err != nil {
		return err
	}
	return nil
}

type SRProductSlice []SRProducts

func (SRProductSlice) GormDataType() string {
	return "json"
}

func (data SRProductSlice) Value() (driver.Value, error) {
	jsonifiedData, _ := json.Marshal(data)
	return string(jsonifiedData), nil
}

func (data *SRProductSlice) Scan(value interface{}) error {
	err := GormScan(reflect.ValueOf(data), value)
	if err != nil {
		return err
	}
	return nil
}

type SRShipment struct {
	ID                 int     `json:"id"`
	OrderID            int     `json:"order_id"`
	OrderProductID     any     `json:"order_product_id"`
	ChannelID          int     `json:"channel_id"`
	Code               string  `json:"code"`
	Cost               string  `json:"cost"`
	Tax                string  `json:"tax"`
	Awb                any     `json:"awb"`
	RtoAwb             string  `json:"rto_awb"`
	AwbAssignDate      any     `json:"awb_assign_date"`
	Etd                string  `json:"etd"`
	DeliveredDate      string  `json:"delivered_date"`
	Quantity           int     `json:"quantity"`
	CodCharges         string  `json:"cod_charges"`
	Number             any     `json:"number"`
	Name               any     `json:"name"`
	OrderItemID        any     `json:"order_item_id"`
	Weight             float32 `json:"weight"`
	VolumetricWeight   float64 `json:"volumetric_weight"`
	Dimensions         string  `json:"dimensions"`
	Comment            string  `json:"comment"`
	Courier            string  `json:"courier"`
	CourierID          int     `json:"courier_id"`
	ManifestID         string  `json:"manifest_id"`
	ManifestEscalate   bool    `json:"manifest_escalate"`
	Status             string  `json:"status"`
	IsdCode            string  `json:"isd_code"`
	CreatedAt          string  `json:"created_at"`
	UpdatedAt          string  `json:"updated_at"`
	Pod                any     `json:"pod"`
	EwayBillNumber     string  `json:"eway_bill_number"`
	EwayBillDate       any     `json:"eway_bill_date"`
	Length             float32 `json:"length"`
	Breadth            float32 `json:"breadth"`
	Height             float32 `json:"height"`
	RtoInitiatedDate   string  `json:"rto_initiated_date"`
	RtoDeliveredDate   string  `json:"rto_delivered_date"`
	ShippedDate        string  `json:"shipped_date"`
	PackageImages      string  `json:"package_images"`
	IsRto              bool    `json:"is_rto"`
	EwayRequired       bool    `json:"eway_required"`
	InvoiceLink        string  `json:"invoice_link"`
	IsDarkstoreCourier int     `json:"is_darkstore_courier"`
	CourierCustomRule  string  `json:"courier_custom_rule"`
	IsSingleShipment   bool    `json:"is_single_shipment"`
}

func (SRShipment) GormDataType() string {
	return "json"
}

func (data SRShipment) Value() (driver.Value, error) {
	jsonifiedData, _ := json.Marshal(data)
	return string(jsonifiedData), nil
}

func (data *SRShipment) Scan(value interface{}) error {
	err := GormScan(reflect.ValueOf(data), value)
	if err != nil {
		return err
	}
	return nil
}

type SRAWBData struct {
	Awb            string    `json:"awb"`
	AppliedWeight  float32   `json:"applied_weight"`
	ChargedWeight  string    `json:"charged_weight"`
	BilledWeight   string    `json:"billed_weight"`
	RoutingCode    string    `json:"routing_code"`
	RtoRoutingCode string    `json:"rto_routing_code"`
	Charges        SRCharges `json:"charges"`
}

func (SRAWBData) GormDataType() string {
	return "json"
}

func (data SRAWBData) Value() (driver.Value, error) {
	jsonifiedData, _ := json.Marshal(data)
	return string(jsonifiedData), nil
}

func (data *SRAWBData) Scan(value interface{}) error {
	err := GormScan(reflect.ValueOf(data), value)
	if err != nil {
		return err
	}
	return nil
}

func (SRCharges) GormDataType() string {
	return "json"
}

func (data SRCharges) Value() (driver.Value, error) {
	jsonifiedData, _ := json.Marshal(data)
	return string(jsonifiedData), nil
}

func (data *SRCharges) Scan(value interface{}) error {
	err := GormScan(reflect.ValueOf(data), value)
	if err != nil {
		return err
	}
	return nil
}

type SRCharges struct {
	Zone                   string  `json:"zone"`
	CodCharges             float32 `json:"cod_charges"`
	AppliedWeightAmount    string  `json:"applied_weight_amount"`
	FreightCharges         string  `json:"freight_charges"`
	AppliedWeight          string  `json:"applied_weight"`
	ChargedWeight          string  `json:"charged_weight"`
	ChargedWeightAmount    string  `json:"charged_weight_amount"`
	ChargedWeightAmountRto string  `json:"charged_weight_amount_rto"`
	AppliedWeightAmountRto string  `json:"applied_weight_amount_rto"`
	ServiceTypeID          string  `json:"service_type_id"`
}

type SRReturnPickupData struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Address   string `json:"address"`
	Address2  string `json:"address_2"`
	City      string `json:"city"`
	State     string `json:"state"`
	Country   string `json:"country"`
	PinCode   string `json:"pin_code"`
	Phone     string `json:"phone"`
	Lat       any    `json:"lat"`
	Long      any    `json:"long"`
	OrderID   int    `json:"order_id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (SRReturnPickupData) GormDataType() string {
	return "json"
}

func (data SRReturnPickupData) Value() (driver.Value, error) {
	jsonifiedData, _ := json.Marshal(data)
	return string(jsonifiedData), nil
}

func (data *SRReturnPickupData) Scan(value interface{}) error {
	err := GormScan(reflect.ValueOf(data), value)
	if err != nil {
		return err
	}
	return nil
}

type SROthersData struct {
	Weight           string `json:"weight"`
	Quantity         int    `json:"quantity"`
	BuyerPsid        any    `json:"buyer_psid"`
	Dimensions       string `json:"dimensions"`
	APIOrderID       string `json:"api_order_id"`
	CompanyName      string `json:"company_name"`
	CurrencyCode     string `json:"currency_code"`
	PackageCount     string `json:"package_count"`
	ShippingCity     string `json:"shipping_city"`
	ShippingName     string `json:"shipping_name"`
	ShippingEmail    string `json:"shipping_email"`
	ShippingPhone    string `json:"shipping_phone"`
	ShippingState    string `json:"shipping_state"`
	CustomOrderID    any    `json:"custom_order_id"`
	BillingIsdCode   string `json:"billing_isd_code"`
	ForwardOrderID   any    `json:"forward_order_id"`
	ShippingAddress  string `json:"shipping_address"`
	ShippingCharges  string `json:"shipping_charges"`
	ShippingCountry  string `json:"shipping_country"`
	ShippingPincode  string `json:"shipping_pincode"`
	ShippingAddress2 string `json:"shipping_address_2"`
}

func (SROthersData) GormDataType() string {
	return "json"
}

func (data SROthersData) Value() (driver.Value, error) {
	jsonifiedData, _ := json.Marshal(data)
	return string(jsonifiedData), nil
}

func (data *SROthersData) Scan(value interface{}) error {
	err := GormScan(reflect.ValueOf(data), value)
	if err != nil {
		return err
	}
	return nil
}

type SRExtraInfo struct {
	QcCheck                      int    `json:"qc_check"`
	QcParams                     string `json:"qc_params"`
	OrderType                    int    `json:"order_type"`
	AmazonDgStatus               bool   `json:"amazon_dg_status"`
	ForwardOrderID               string `json:"forward_order_id"`
	BluedartDgStatus             bool   `json:"bluedart_dg_status"`
	OtherCourierDgStatus         bool   `json:"other_courier_dg_status"`
	InsuraceOptedAtOrderCreation bool   `json:"insurace_opted_at_order_creation"`
}

func (SRExtraInfo) GormDataType() string {
	return "json"
}

func (data SRExtraInfo) Value() (driver.Value, error) {
	jsonifiedData, _ := json.Marshal(data)
	return string(jsonifiedData), nil
}

func (data *SRExtraInfo) Scan(value interface{}) error {
	err := GormScan(reflect.ValueOf(data), value)
	if err != nil {
		return err
	}
	return nil
}

type SRRefundDetail struct {
	RefundMode        string `json:"refund_mode"`
	AccountHolderName string `json:"account_holder_name"`
	AccountNumber     string `json:"account_number"`
	BankIfsc          string `json:"bank_ifsc"`
	BankName          string `json:"bank_name"`
}

func (SRRefundDetail) GormDataType() string {
	return "json"
}

func (data SRRefundDetail) Value() (driver.Value, error) {
	jsonifiedData, _ := json.Marshal(data)
	return string(jsonifiedData), nil
}

func (data *SRRefundDetail) Scan(value interface{}) error {
	err := GormScan(reflect.ValueOf(data), value)
	if err != nil {
		return err
	}
	return nil
}

func MigrateShiprocketOrderToMySQL(fullShiprocketOrder FullShiprocketOrder) (err error) {
	err = Mysql.DB.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&fullShiprocketOrder).Error

	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

func GetShiprocketOrderByChannelOrderID(orderID string) (FullShiprocketOrder, error) {
	var fullShiprocketOrder FullShiprocketOrder
	err := db.ShiprocketOrderCollection.FindOne(context.Background(), bson.M{"channelorderid": orderID}).Decode(&fullShiprocketOrder)
	if err != nil {
		return fullShiprocketOrder, err
	}

	return fullShiprocketOrder, nil
}

func IsOrderPaymentRemitted(order Order) (bool, error) {
	if order.Payment.Method != "COD" {
		return true, nil
	}

	shiprocketOrder, err := GetShiprocketOrderByChannelOrderID(order.ID)
	if err != nil {
		return false, err
	}

	if shiprocketOrder.RemittanceStatus == "Remittance success" {
		return true, nil
	} else {
		if len(shiprocketOrder.RemittanceStatus) > 0 {
			return false, fmt.Errorf("remittance status is: %s for order id: %s", shiprocketOrder.RemittanceStatus, order.ID)
		}
	}
	return false, nil
}
