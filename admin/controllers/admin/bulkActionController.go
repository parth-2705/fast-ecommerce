package controllers

import (
	"context"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"hermes/configs/Redis"
	"hermes/db"
	"hermes/models"
	"hermes/search"
	"hermes/utils/data"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type InsertCatalogueRequestBody struct {
	Seller  string         `json:"seller"`
	Brand   string         `json:"brand"`
	Mapping map[string]CSV `json:"mapping"`
}

func GetVariantAttributes(record []string, inverseMapping map[string]CSV) (map[string]string, error) {
	attributes := make(map[string]string, 0)

	nameKey := "variantName"
	valueKey := "variantValue"

	for i := 1; i <= 4; i++ {
		nameRecordIndex := inverseMapping[nameKey+fmt.Sprint(i)].Index
		valueRecordIndex := inverseMapping[valueKey+fmt.Sprint(i)].Index

		if nameRecordIndex >= len(record) || nameRecordIndex < 0 {
			return attributes, fmt.Errorf("name index out of bounds")
		}

		if valueRecordIndex >= len(record) || valueRecordIndex < 0 {
			return attributes, fmt.Errorf("value index out of bounds")
		}

		if len(record[nameRecordIndex]) > 0 && len(record[valueRecordIndex]) > 0 {
			attributes[record[nameRecordIndex]] = record[valueRecordIndex]
		}
	}
	return attributes, nil
}

func GetProductAttributes(records [][]string, inverseMapping map[string]CSV) ([]models.VariationAttribute, error) {
	attributes := make([]models.VariationAttribute, 0)

	nameKey := "variantName"
	valueKey := "variantValue"

	for i := 1; i <= 4; i++ {
		nameRecordIndex := inverseMapping[nameKey+fmt.Sprint(i)].Index
		valueRecordIndex := inverseMapping[valueKey+fmt.Sprint(i)].Index

		attribute := models.VariationAttribute{}

		for _, record := range records {
			if nameRecordIndex >= len(record) || nameRecordIndex < 0 {
				return attributes, fmt.Errorf("name index out of bounds")
			}

			if valueRecordIndex >= len(record) || valueRecordIndex < 0 {
				return attributes, fmt.Errorf("value index out of bounds")
			}

			if len(record[nameRecordIndex]) > 0 && len(record[valueRecordIndex]) > 0 {
				attribute.Name = record[nameRecordIndex]
				attribute.VisType = 0
				if record[nameRecordIndex] == "Color" || record[nameRecordIndex] == "color" {
					attribute.VisType = 1
				}
				attribute.Options = append(attribute.Options, record[valueRecordIndex])
			}
		}

		if len(attribute.Name) > 0 {
			attributes = append(attributes, attribute)
		}

	}
	return attributes, nil
}

func ProcessSpecification(inverseMapping map[string]CSV, record []string) ([]models.KV, error) {
	var specifications []models.KV

	optionKey := "option"
	valueKey := "value"

	for i := 1; i <= 4; i++ {
		optionRecordIndex := inverseMapping[optionKey+fmt.Sprint(i)].Index
		valueRecordIndex := inverseMapping[valueKey+fmt.Sprint(i)].Index

		if optionRecordIndex >= len(record) || optionRecordIndex < 0 {
			return specifications, fmt.Errorf("option index out of bounds")
		}

		if valueRecordIndex >= len(record) || valueRecordIndex < 0 {
			return specifications, fmt.Errorf("value index out of bounds")
		}

		if len(record[optionRecordIndex]) > 0 && len(record[valueRecordIndex]) > 0 {
			specifications = append(specifications, models.KV{
				Key:   record[optionRecordIndex],
				Value: record[valueRecordIndex],
			})
		}
	}
	return specifications, nil
}

func ProcessImagesFromMapping(inverseMapping map[string]CSV, record []string) ([]models.MediaObject, error) {
	baseImageURL := "imageUrl"
	var media []models.MediaObject

	for i := 1; i <= 10; i++ {
		key := baseImageURL + fmt.Sprint(i)
		imageRecordIndex := inverseMapping[key].Index

		if imageRecordIndex >= len(record) || imageRecordIndex < 0 {
			return media, fmt.Errorf("image index out of bounds")
		}

		if len(record[imageRecordIndex]) > 0 {
			media = append(media, models.MediaObject{
				ID:   record[imageRecordIndex],
				Type: models.TypeImage,
			})
		}
	}

	if len(media) < 1 {
		return media, fmt.Errorf("no images found")
	}

	return media, nil
}

func UpdateKeyInVariant(variant *models.Variation, key string, value string) error {
	switch key {
	case "price.sellingPrice":
		if sellingPrice, err := strconv.ParseFloat(value, 32); err == nil {
			variant.Price.SellingPrice = sellingPrice
			variant.Price.Discount = variant.Price.MRP - variant.Price.SellingPrice
			variant.Price.DiscountPercentage = ((variant.Price.MRP - variant.Price.SellingPrice) / variant.Price.MRP) * 100
		} else {
			return err
		}
	case "price.sellerPrice":
		fmt.Println("variant.Price.SellerPrice", value)
		if sellerPrice, err := strconv.ParseFloat(value, 32); err == nil {
			variant.Price.SellerPrice = sellerPrice
		} else {
			return err
		}
	case "price.mrp":
		if mrp, err := strconv.ParseFloat(value, 32); err == nil {
			variant.Price.MRP = mrp
			variant.Price.Discount = variant.Price.MRP - variant.Price.SellingPrice
			variant.Price.DiscountPercentage = ((variant.Price.MRP - variant.Price.SellingPrice) / variant.Price.MRP) * 100
		} else {
			return err
		}
	case "inventory":
		if quantity, err := strconv.Atoi(value); err == nil {
			variant.Quantity = quantity
		} else {
			return err
		}
	case "status":
		variant.Status = value
	case "weight":
		if len(value) == 0 {
			return fmt.Errorf("empty weight")
		}
		variant.Weight = value
	case "length":
		if len(value) == 0 {
			return fmt.Errorf("empty length")
		}
		variant.Length = value
	case "breadth":
		if len(value) == 0 {
			return fmt.Errorf("empty breadth")
		}
		variant.Breadth = value
	case "height":
		if len(value) == 0 {
			return fmt.Errorf("empty height")
		}
		variant.Height = value
	case "sku":
		if len(value) == 0 {
			return fmt.Errorf("empty sku")
		}
		variant.SKU = value
	case "hsn":
		if len(value) == 0 {
			return fmt.Errorf("empty hsn")
		}
		variant.Barcode = value
	default:
		return nil
	}
	return nil
}

func UpdateKeyInProduct(product *models.Product, key string, value string) error {
	switch key {
	case "title":
		if len(value) == 0 {
			return fmt.Errorf("empty title")
		}
		product.Name = value
	case "description":
		if len(value) == 0 {
			return fmt.Errorf("empty description")
		}
		encodedText := base64.StdEncoding.EncodeToString([]byte(value))
		product.DescriptionEncoded = encodedText
	case "gender":
		product.Gender = value
	case "category":
		if len(value) == 0 {
			return fmt.Errorf("empty category")
		}
		product.Category = value
	case "subCategory":
		if len(value) == 0 {
			return fmt.Errorf("empty subcategory")
		}
		product.SubCategory = value
	case "productType":
		if len(value) == 0 {
			return fmt.Errorf("empty product type")
		}
		product.ProductType = value
	case "categoryID":
		if len(value) == 0 {
			return fmt.Errorf("empty category ID")
		}
		product.NewCategory = value
	case "shelfLifeDays":
		if len(value) == 0 {
			return fmt.Errorf("empty shelf life days")
		}
		shelfLife, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		product.ShelfLifeDays = shelfLife
	case "isBestSeller":
		product.IsBestSeller = value == "Yes"
	case "isHotDeal":
		product.IsHotDeal = value == "Yes"
	case "ean":
		product.EANCode = value
	case "gst":
		if len(value) == 0 {
			return fmt.Errorf("empty gst")
		}
		product.GST = strings.Replace(value, "%", "", -1)
	case "cess":
		product.Cess = strings.Replace(value, "%", "", -1)
	case "price.sellingPrice":
		if sellingPrice, err := strconv.ParseFloat(value, 32); err == nil {
			product.Price.SellingPrice = sellingPrice
			product.Price.Discount = product.Price.MRP - product.Price.SellingPrice
			product.Price.DiscountPercentage = ((product.Price.MRP - product.Price.SellingPrice) / product.Price.MRP) * 100
		} else {
			return err
		}
	case "price.mrp":
		if mrp, err := strconv.ParseFloat(value, 32); err == nil {
			product.Price.MRP = mrp
			product.Price.Discount = product.Price.MRP - product.Price.SellingPrice
			product.Price.DiscountPercentage = ((product.Price.MRP - product.Price.SellingPrice) / product.Price.MRP) * 100
		} else {
			return err
		}
	case "price.sellerPrice":
		fmt.Println("product.Price.SellerPrice", value)
		if sellerPrice, err := strconv.ParseFloat(value, 32); err == nil {
			product.Price.SellerPrice = sellerPrice
		} else {
			return err
		}
	default:
		return nil
	}
	return nil
}

func CheckIfSKUIsPresentOrNot(keys []reflect.Value, requestBody map[string]CSV) bool {
	skuPresent := false
	for _, keyReflect := range keys {
		key := keyReflect.Interface().(string)
		toUpdate := requestBody[key]

		if toUpdate.Value == "sku" {
			skuPresent = true
			break
		}
	}
	return skuPresent
}

func CheckIfMandatoryKeysIsPresentOrNot(keys []reflect.Value, requestBody map[string]CSV) (bool, string) {
	return CheckIfSKUIsPresentOrNot(keys, requestBody), "SKU ID"
}

func GetCSVFromURL(url string) ([][]string, error) {
	var content [][]string
	httpRequest, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return content, err
	}

	client := http.Client{Timeout: time.Second * 10}
	response, err := client.Do(httpRequest)
	if err != nil {
		return content, err
	}

	if response.StatusCode != http.StatusOK {
		return content, fmt.Errorf("response with status code: %d", response.StatusCode)
	}

	content, err = csv.NewReader(response.Body).ReadAll()
	if err != nil {
		return content, err
	}
	return content, err
}

func GetSKUsFilterFromRequest(keys []reflect.Value, requestBody map[string]CSV, header []string, record []string) (bson.M, error) {
	filter := bson.M{}
	for _, keyReflect := range keys {
		key := keyReflect.Interface().(string)
		toUpdate := requestBody[key]

		if toUpdate.Index < 0 || toUpdate.Index >= len(header) {
			return filter, fmt.Errorf("index out of bounds")
		}

		if toUpdate.Value == "sku" {
			filter["sku"] = record[toUpdate.Index]
		}
	}
	return filter, nil
}

func BulkUpdate(c *gin.Context) {
	csvURL := c.Query("url")

	var requestBody map[string]CSV
	if err := c.ShouldBind(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	keys := reflect.ValueOf(requestBody).MapKeys()

	// one column is for SKU
	if len(keys) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No Columns to update"})
		return
	}

	// Check if SKU is present in the request body or not
	skuPresent := CheckIfSKUIsPresentOrNot(keys, requestBody)

	// If not present return an error
	if !skuPresent {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SKU ID not present or mapped"})
		return
	}

	// Download the csv file from bucket url
	content, err := GetCSVFromURL(csvURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// First row is header
	header := content[0]
	records := content[1:]

	session, err := db.MongoClient.StartSession()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	defer session.EndSession(context.Background())

	// Starting transaction
	err = session.StartTransaction()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var productsBeforeChange []models.Product
	var variantsBeforeChange []models.Variation

	// Read the records
	for _, record := range records {

		// Get the SKU filter
		filter, err := GetSKUsFilterFromRequest(keys, requestBody, header, record)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			session.AbortTransaction(context.Background())
			return
		}

		// Find the variant
		var variant models.Variation
		opts := options.FindOne().SetSort(bson.D{{Key: "updatedAt", Value: -1}})
		err = db.VariantsCollection.FindOne(context.Background(), filter, opts).Decode(&variant)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				fmt.Println("not found sku: ", fmt.Sprint(filter["sku"]))
				continue
			}

			c.JSON(http.StatusBadRequest, gin.H{"error": "while finding variant " + err.Error() + " sku: " + fmt.Sprint(filter["sku"])})
			return
		}

		// Find the product
		var product models.Product
		err = db.ProductCollection.FindOne(context.Background(), bson.M{"_id": variant.ProductID}).Decode(&product)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "while finding product " + err.Error() + " of variant: " + variant.ProductID})
			return
		}

		for _, keyReflect := range keys {
			key := keyReflect.Interface().(string)
			toUpdate := requestBody[key]

			if toUpdate.Value != "sku" {
				err := UpdateKeyInVariant(&variant, toUpdate.Value, record[toUpdate.Index])
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "inserting key in variant" + err.Error()})
					session.AbortTransaction(context.Background())
					return
				}

				err = UpdateKeyInProduct(&product, toUpdate.Value, record[toUpdate.Index])
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "while updating key in product " + err.Error()})
					session.AbortTransaction(context.Background())
					return
				}
			}
		}

		variant.UpdatedAt = time.Now()
		product.UpdatedAt = time.Now()

		var oldVariant models.Variation
		err = db.VariantsCollection.FindOneAndUpdate(context.Background(), filter, bson.M{"$set": variant}).Decode(&oldVariant)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "find and update varaint " + err.Error(), "filter": filter, "variant": variant})
			session.AbortTransaction(context.Background())
			return
		}
		variantsBeforeChange = append(variantsBeforeChange, oldVariant)

		var oldProduct models.Product
		err = db.ProductCollection.FindOneAndUpdate(context.Background(), bson.M{"_id": variant.ProductID}, bson.M{"$set": product}).Decode(&oldProduct)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "find and update product " + err.Error(), "filter": variant.ProductID, "variant": product})
			session.AbortTransaction(context.Background())
			return
		}
		productsBeforeChange = append(productsBeforeChange, oldProduct)

		err = Redis.DeleteProductCacheByID(oldProduct.ID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "find and update product " + err.Error()})
			session.AbortTransaction(context.Background())
			return
		}

		// Update product in mille
		go search.UpdateProductByID(oldProduct.ID)
	}

	// Commit the transaction
	err = session.CommitTransaction(context.Background())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	} else {
		var bulkAction models.BulkAction
		bulkAction.ID = data.GetUUIDString("bulkAction")
		bulkAction.CSV = csvURL

		user := c.Request.Header.Get("user")
		bulkAction.ActionTakenBy = user
		version, err := models.GetBulkActionVersion()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Not able to retrieve version number"})
			return
		}
		bulkAction.Version = int(version + 1)
		bulkAction.ProductsBeforeChange = productsBeforeChange
		bulkAction.VariantsBeforeChange = variantsBeforeChange

		err = bulkAction.Create()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"response": "Success"})
}

func UploadProductsAndItsVariants(requestBody InsertCatalogueRequestBody, brand models.Brand, inverseMapping map[string]CSV, rows [][]string, keys []reflect.Value) ([]interface{}, []interface{}, error) {
	var product models.Product
	var variant models.Variation

	// Read the records
	var products []interface{}
	var variants []interface{}

	var currentProductVariants []models.Variation

	product.ID = data.GetUUIDString("product")
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()
	product.SellerID = requestBody.Seller
	product.BrandID = requestBody.Brand
	product.Brand = brand

	media, err := ProcessImagesFromMapping(inverseMapping, rows[0])
	if err != nil {
		return products, variants, err
	}

	product.Media = media
	product.Thumbnail = models.Image(media[0].ID)

	specifications, err := ProcessSpecification(inverseMapping, rows[0])
	if err != nil {
		return products, variants, err
	}
	product.Specifications = specifications

	for _, keyReflect := range keys {
		key := keyReflect.Interface().(string)
		toUpdate := requestBody.Mapping[key]

		err := UpdateKeyInProduct(&product, toUpdate.Value, rows[0][toUpdate.Index])
		if err != nil {
			return products, variants, err
		}
	}

	productAttributes, err := GetProductAttributes(rows, inverseMapping)
	if err != nil {
		return products, variants, err
	}
	product.Attributes = productAttributes

	// Add all payment methods to product
	product.PaymentMethods = make(models.PaymentMethodMap)
	for _, po := range models.CopyPaymentOptions() {
		product.PaymentMethods[po.ID] = models.PaymentMethodConfiguration{Available: true}
	}

	// To be inserted in mongodb
	products = append(products, product)

	for _, variantRecord := range rows {

		variant.ID = data.GetUUIDString("variant")
		variant.ProductID = product.ID
		variant.CreatedAt = time.Now()
		variant.UpdatedAt = time.Now()

		variant.Media = media
		variant.Thumbnail = models.Image(media[0].ID)

		for _, keyReflect := range keys {
			key := keyReflect.Interface().(string)
			toUpdate := requestBody.Mapping[key]

			err = UpdateKeyInVariant(&variant, toUpdate.Value, variantRecord[toUpdate.Index])
			if err != nil {
				return products, variants, err
			}
		}

		attributes, err := GetVariantAttributes(variantRecord, inverseMapping)
		if err != nil {
			return products, variants, err
		}

		variant.Attributes = attributes

		// To be inserted in mongodb
		variants = append(variants, variant)

		// To handle current product variants
		currentProductVariants = append(currentProductVariants, variant)
	}

	productVariants := make([]models.Variation, len(currentProductVariants))
	copy(productVariants, currentProductVariants)
	product.Variants = productVariants
	return products, variants, nil
}

func InsertCatalogue(c *gin.Context) {
	csvURL := c.Query("url")

	var requestBody InsertCatalogueRequestBody
	if err := c.ShouldBind(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	keys := reflect.ValueOf(requestBody.Mapping).MapKeys()

	// one column is for SKU
	if len(keys) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No Columns to update"})
		return
	}

	// Check if SKU is present in the request body or not
	arePresent, key := CheckIfMandatoryKeysIsPresentOrNot(keys, requestBody.Mapping)

	// If not present return an error
	if !arePresent {
		c.JSON(http.StatusBadRequest, gin.H{"error": key + " not present or mapped"})
		return
	}

	// Download the csv file from bucket url
	content, err := GetCSVFromURL(csvURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// First row is header
	records := content[1:]

	session, err := db.MongoClient.StartSession()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	defer session.EndSession(context.Background())

	// Starting transaction
	err = session.StartTransaction()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	inverseMapping := make(map[string]CSV, 0)
	for _, keyReflext := range keys {
		key := keyReflext.Interface().(string)
		toInverse := requestBody.Mapping[key]
		inverseMapping[toInverse.Value] = CSV{
			Index: toInverse.Index,
			Value: key,
		}
	}

	brand, err := models.GetBrand(requestBody.Brand)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		session.AbortTransaction(context.Background())
		return
	}

	titleIndex := inverseMapping["title"].Index
	if titleIndex >= len(content[0]) || titleIndex < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Errorf("title index out of bounds")})
		session.AbortTransaction(context.Background())
		return
	}

	var rows [][]string

	// Read the records
	var products []interface{}
	var variants []interface{}

	totalRecordsFound := len(records)
	for _, record := range records {

		if len(rows) == 0 {
			rows = append(rows, record)
		} else {
			if record[titleIndex] == rows[len(rows)-1][titleIndex] {
				rows = append(rows, record)
			} else {
				currentProducts, currentVariants, err := UploadProductsAndItsVariants(requestBody, brand, inverseMapping, rows, keys)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					session.AbortTransaction(context.Background())
					return
				}

				products = append(products, currentProducts...)
				variants = append(variants, currentVariants...)

				rows = make([][]string, 0)
				rows = append(rows, record)
			}

		}

	}

	if len(rows) > 0 {
		currentProducts, currentVariants, err := UploadProductsAndItsVariants(requestBody, brand, inverseMapping, rows, keys)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			session.AbortTransaction(context.Background())
			return
		}

		products = append(products, currentProducts...)
		variants = append(variants, currentVariants...)

	}

	_, err = db.ProductCollection.InsertMany(context.Background(), products)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "while inserting products data " + err.Error()})
		session.AbortTransaction(context.Background())
		return
	}

	_, err = db.VariantsCollection.InsertMany(context.Background(), variants)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "while inserting variants data " + err.Error()})
		session.AbortTransaction(context.Background())
		return
	}

	go search.AddCatalogToIndex(products)

	// Commit the transaction
	err = session.CommitTransaction(context.Background())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": "Success", "records": totalRecordsFound})
}
