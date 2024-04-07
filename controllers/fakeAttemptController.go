package controllers

import (
	"fmt"
	"hermes/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func FakeAttemptController(c *gin.Context) {

	awb := c.Query("awb")

	if len(awb) == 0 {
		fmt.Println("no awb found")
		c.Redirect(http.StatusFound, "/")
		return
	}

	shipment, err := models.GetShipmentByAWB(awb)
	if err != nil {
		fmt.Println("shipment error: ", err)
		c.Redirect(http.StatusFound, "/")
		return
	}

	order, err := models.GetOrder(shipment.ParentOrderID)
	if err != nil {
		fmt.Println("order error: ", err)
		c.Redirect(http.StatusFound, "/")
		return
	}

	currentTime := time.Now().Format("15:04")

	paymentMethod := "Cash on Delivery"
	if order.Payment.Method != "COD" {
		paymentMethod = "Prepaid"
	}

	c.HTML(http.StatusOK, "fake-attempt", gin.H{"order": order, "time": currentTime, "paymentMethod": paymentMethod})
}

func DownloadFakeAttemptReportV2(c *gin.Context) {

	awb := c.Query("awb")

	if len(awb) == 0 {
		fmt.Println("no awb found")
		c.Redirect(http.StatusFound, "/")
		return
	}

	shipment, err := models.GetShipmentByAWB(awb)
	if err != nil {
		fmt.Println("shipment error: ", err)
		c.Redirect(http.StatusFound, "/")
		return
	}

	order, err := models.GetOrder(shipment.ParentOrderID)
	if err != nil {
		fmt.Println("order error: ", err)
		c.Redirect(http.StatusFound, "/")
		return
	}

	currentTime := time.Now().Format("15:04")

	paymentMethod := "Cash on Delivery"
	if order.Payment.Method != "COD" {
		paymentMethod = "Prepaid"
	}

	// Create a capturer to intercept the response
	// capturer := &ResponseCapturer{
	// 	ResponseWriter: c.Writer,
	// 	body:           &bytes.Buffer{},
	// }
	// c.Writer = capturer

	// Serve the HTML page using Gin
	c.HTML(http.StatusOK, "fake-attempt", gin.H{"order": order, "time": currentTime, "paymentMethod": paymentMethod})

	// Get the captured HTML content
	// html := capturer.body.String()

	// err = GeneratePDF(html, "./output.pdf")
	// if err != nil {
	// 	// Handle the error
	// 	fmt.Println("Example error: ", err)
	// 	return
	// }

	// Convert HTML to an image using wkhtmltoimage
	// pdfg, err := wkhtmltopdf.NewPDFGenerator()
	// if err != nil {
	// 	// Handle the error
	// 	fmt.Println("wkhtmltopdf error: ", err)
	// 	return
	// }
	// pdfg.AddPage(wkhtmltopdf.NewPageReader(strings.NewReader(html)))
	// err = pdfg.Create()
	// if err != nil {
	// 	// Handle the error
	// 	fmt.Println("wkhtmltopdf AddPage error: ", err)
	// 	return
	// }

	// // Save the image to a file
	// err = pdfg.WriteFile("output.png")
	// if err != nil {
	// 	// Handle the error
	// 	fmt.Println("wkhtmltopdf WriteFile error: ", err)
	// 	return
	// }

	// Respond with a download link to the image
	// c.Header("Content-Disposition", "attachment; filename=output.png")
	// c.File("output.png")
}
