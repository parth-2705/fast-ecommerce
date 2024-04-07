package shiprocket

import (
	"encoding/json"
	"fmt"
	"hermes/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tryamigo/themis"
	"go.mongodb.org/mongo-driver/bson"
)

type ShipRocketPickUpForShipment struct {
	ShipmentID []int    `json:"shipment_id,omitempty"` //required
	PickUpDate []string `json:"pickup_date,omitempty"`
}

type AlreadyInPickUpQueue struct {
	Message     string `json:"message"`
	Status_Code int    `json:"status_code"`
}

func UpdateShipmentStatusDatabase(shipmentId int, shipmentStatus bool, manifestUrl string) error {
	var shipping models.Shipping
	shipping.ShippingId = shipmentId

	toUpdateFields := bson.M{"shipmentCreated": shipmentStatus, "manifestUrl": manifestUrl}
	err := shipping.UpdateFields(toUpdateFields)
	if err != nil {
		return err
	}

	return nil
}

func CreateShipRocketPickUpRequestBody(shipmentId int, date string) ShipRocketPickUpForShipment {
	var ShipRocketPickUpForShipment ShipRocketPickUpForShipment
	ShipRocketPickUpForShipment.ShipmentID = append(ShipRocketPickUpForShipment.ShipmentID, shipmentId)

	if len(date) > 0 {
		ShipRocketPickUpForShipment.PickUpDate = append(ShipRocketPickUpForShipment.PickUpDate, date)
	}

	return ShipRocketPickUpForShipment
}

func CreateShipRocketPickUpUtil(requestBody []byte, headers [][]string) error {

	url := "https://apiv2.shiprocket.in/v1/external/courier/generate/pickup"
	statusCode, responseBody, err := themis.HitAPIEndpoint2(url, http.MethodPost, requestBody, headers, [][]string{})
	if err != nil {
		return err
	}
	if statusCode >= 400 {
		if statusCode == 400 {
			var alreadyInPickUpQueue AlreadyInPickUpQueue
			err = json.Unmarshal(responseBody, &alreadyInPickUpQueue)
			if err != nil {
				return fmt.Errorf("erorr: %s with status code: %d", string(responseBody), statusCode)
			}

			if alreadyInPickUpQueue.Status_Code != 400 || alreadyInPickUpQueue.Message != "Already in Pickup Queue." {
				return fmt.Errorf("erorr: %s with status code: %d", string(responseBody), statusCode)
			}
		} else {
			return fmt.Errorf("erorr: %s with status code: %d", string(responseBody), statusCode)
		}
	}
	return nil
}

func CreateShipRocketPickUp(shipmentId int, date string) error {

	token, err := GetShipRocketToken()
	if err != nil {
		return err
	}

	headers := [][]string{{"Authorization", "Bearer " + token}}

	shipRocketPickUpForShipment := CreateShipRocketPickUpRequestBody(shipmentId, date)
	requestBody, _ := json.Marshal(shipRocketPickUpForShipment)

	err = CreateShipRocketPickUpUtil(requestBody, headers)
	if err != nil {
		return err
	}

	shipmentStatus := true
	manifestUrl, err := CreateShipRocketShipmentManifest(fmt.Sprint(shipmentId), headers)
	if err != nil {

		// update pickup in shipment and let the manifest url be empty
		_ = UpdateShipmentStatusDatabase(shipmentId, shipmentStatus, "")
		return err
	}

	err = UpdateShipmentStatusDatabase(shipmentId, shipmentStatus, manifestUrl)
	if err != nil {
		return err
	}

	return nil
}

func CreatePickupOnShipRocket(c *gin.Context) {
	shipmentID := c.Query("id")
	if len(shipmentID) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Shipment id not found"})
		return
	}

	date := c.Query("date")

	id, err := strconv.Atoi(shipmentID)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("couldn't process shipment id %s", err.Error())})
		fmt.Printf("err: %v\n", err)
		return
	}

	err = CreateShipRocketPickUp(id, date)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("couldn't process order pickup %s", err.Error())})
		fmt.Printf("err: %v\n", err)
		return
	}

	shipment, err := models.GetShipmentByShipmentID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("couldn't find shipment %s", err.Error())})
		fmt.Printf("err: %v\n", err)
		return
	}

	if shipment.ParentRelation != 0 {
		err = models.MarkStatusForBackwardShipment(shipment.Id, "PICKUP GENERATED")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("couldn't mark backward shipment aspickup generated %s", err.Error())})
			fmt.Printf("err: %v\n", err)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"shipment": shipment})
}

func GenerateOrderManifest(c *gin.Context) {
	shipmentID := c.Query("id")
	if len(shipmentID) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Shipment id not found"})
		return
	}

	token, err := GetShipRocketToken()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	headers := [][]string{{"Authorization", "Bearer " + token}}

	manifestUrl, err := CreateShipRocketShipmentManifest(shipmentID, headers)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := strconv.Atoi(shipmentID)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("couldn't process shipment id %s", err.Error())})
		fmt.Printf("err: %v\n", err)
		return
	}

	err = UpdateShipmentStatusDatabase(id, true, manifestUrl)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": "Success", "manifestUrl": manifestUrl})
}
