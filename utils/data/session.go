package data

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"hermes/configs"
	"net"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func AddSessionFlashMessage(c *gin.Context, message string) {
	sessions.Default(c).AddFlash(message)
	sessions.Default(c).Save()
}

func GetSessionFlashMessages(c *gin.Context) []interface{} {
	flashes := sessions.Default(c).Flashes()
	sessions.Default(c).Save()
	return flashes
}

func SetSessionValue(c *gin.Context, key string, value interface{}) {
	sessions.Default(c).Set(key, value)
	sessions.Default(c).Save()
}

func GetSessionValue(c *gin.Context, key string) interface{} {
	return sessions.Default(c).Get(key)
}

func GetProductIDFromSession(c *gin.Context) (productID string) {
	productIDInterface := GetSessionValue(c, configs.ProductKey)
	productID, _ = productIDInterface.(string)
	return
}

func SetCouponInSession(c *gin.Context, couponCode string) {
	SetSessionValue(c, configs.CouponKey, couponCode)
}

func SetCartInSession(c *gin.Context, cartID string) {
	SetSessionValue(c, configs.Cart, cartID)
}

func GetCartFromSession(c *gin.Context) string {
	cartID, _ := GetSessionValue(c, configs.Cart).(string)
	return cartID
}

// func RemoveCartFromSession(c *gin.Context) {
// 	SetSessionValue(c, configs.Cart, nil)
// }

func GetUserAgentIDFromSession(c *gin.Context) string {
	val, ok := GetSessionValue(c, configs.UserAgentIdentifier).(string)
	if !ok {
		return ""
	}
	return val
}

func GetUserIDFromSession(c *gin.Context) string {

	userInterface := GetSessionValue(c, configs.Userkey)
	if user, ok := userInterface.(string); !ok {
		return ""
	} else {
		return user
	}
}

func SetInternalUserInSession(c *gin.Context, internal bool) {
	SetSessionValue(c, configs.Internal, internal)
}

func GetIfUserIsInternalFromSession(c *gin.Context) (internal bool) {
	internalInterface := GetSessionValue(c, configs.Internal)
	internal, _ = internalInterface.(bool)
	return
}

func SetUserAgentInSession(c *gin.Context, userAgent string) {
	SetSessionValue(c, configs.UserAgent, userAgent)
}

func GetUserAgentFromSession(c *gin.Context) (userAgent string) {
	return c.Request.UserAgent()
}

func SetIPAddressinSession(c *gin.Context, ipAddress string) {
	SetSessionValue(c, configs.IP, ipAddress)
}

func GetIPAddressFromSession(c *gin.Context) (ipAddress string) {
	realIP := c.Request.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	lastIP, _, _ := net.SplitHostPort(c.Request.RemoteAddr)
	return lastIP
}

func GetFBPCookie(c *gin.Context) (fbp string) {
	fbp, _ = c.Cookie("_fbp")
	return
}

func GetFBCCookie(c *gin.Context) (fbc string) {
	fbc, _ = c.Cookie("_fbc")
	return
}

type UTMParams struct {
	Source  string `json:"utm_source" bson:"utm_source"`
	Medium  string `json:"utm_medium" bson:"utm_medium"`
	Name    string `json:"utm_campaign" bson:"utm_campaign"`
	Term    string `json:"utm_term" bson:"utm_term"`
	Content string `json:"utm_content" bson:"utm_content"`
}

func (UTMParams) GormDataType() string {
	return "json"
}

func (data UTMParams) Value() (driver.Value, error) {
	jsonifiedData, _ := json.Marshal(data)
	return string(jsonifiedData), nil
}

func (data *UTMParams) Scan(value interface{}) error {

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

	err := json.Unmarshal(byteSlice, &data)
	return err
}

func SetUTMParamsInSession(c *gin.Context) {

	_, ok := c.GetQuery("utm_source")
	if !ok {
		return
	}

	utmParams := UTMParams{
		Source:  c.Query("utm_source"),
		Medium:  c.Query("utm_medium"),
		Name:    c.Query("utm_campaign"),
		Term:    c.Query("utm_term"),
		Content: c.Query("utm_content"),
	}

	SetSessionValue(c, configs.UTM, utmParams)
}

func GetUTMParamsFromSession(c *gin.Context) UTMParams {
	utmParams, _ := GetSessionValue(c, configs.UTM).(UTMParams)
	return utmParams
}
