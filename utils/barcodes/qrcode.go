package barcodes

import (
	"fmt"
	"image/png"
	"os"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
)

func makeQrCode(qrStr string) (barcode.Barcode, error) {
	qrCode, err := qr.Encode(qrStr, qr.M, qr.Auto)
	if err != nil {
		fmt.Printf("QR err: %v\n", err)
		return qrCode, err
	}

	qrCode, err = barcode.Scale(qrCode, 200, 200)
	if err != nil {
		fmt.Printf("QR Scale err: %v\n", err)
		return qrCode, err
	}

	return qrCode, nil
}

func MakeQRCodeAndStoreOnDisk(qrStr string, fileName string) (err error) {

	qrCode, err := makeQrCode(qrStr)
	if err != nil {
		return
	}

	file, err := os.Create(fileName)
	if err != nil {
		fmt.Printf("QR Code storing err: %v\n", err)
		return
	}

	// encode the barcode as png
	err = png.Encode(file, qrCode)
	if err != nil {
		fmt.Printf("png writing err: %v\n", err)
		return
	}

	file.Close()
	return nil
}
