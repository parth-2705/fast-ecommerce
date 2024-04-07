package db

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/bson"
)

type Pincode struct {
	// pincode is indexed
	Pincode string `bson:"pincode" json:"pincode" index:"{pincode:1}"`
	City    string `bson:"city" json:"city"`
	State   string `bson:"state" json:"state"`
}

func PopulatePincodes() {

	fmt.Println("Populating the 'pincodes' collection of 'mydb' database...")

	// Open the CSV file
	file, err := os.Open("./pincodes2.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Parse the CSV file
	reader := csv.NewReader(file)
	var pinCodes []interface{}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		// Extract the required fields
		pinCode := Pincode{
			Pincode: record[1],
			City:    record[8],
			State:   record[9],
		}

		// Append the Pincode object to the pinCodes slice
		pinCodes = append(pinCodes, pinCode)
	}

	_, err = PincodeCollection.InsertMany(context.Background(), pinCodes)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Pincodes has been successfully inserted.")

}

func PincodesInit() {
	// check if the 'pincodes' collection has any data
	count, err := PincodeCollection.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}

	// if the 'pincodes' collection is empty, populate it
	if count == 0 {
		PopulatePincodes()
	}

	return
}
