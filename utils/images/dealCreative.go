package images

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
)

func DealImageWithQR(barcodePath string) error {

	// Get the Creative (currently present in assets) Should be User Input next
	f, err := os.Open("dealCreative")
	if err != nil {
		return err
	}
	defer f.Close()
	creative, err := png.Decode(f)
	if err != nil {
		return err
	}

	// Get the barcode
	f2, err := os.Open(barcodePath)
	if err != nil {
		return err
	}
	defer f2.Close()
	barcode, err := png.Decode(f2)
	if err != nil {
		return err
	}

	bgImg := image.NewRGBA(image.Rect(0, 0, 1080, 1080))
	draw.Draw(bgImg, bgImg.Bounds(), &image.Uniform{color.RGBA{227, 221, 221, 1}}, image.ZP, draw.Src)

	draw.Draw(bgImg, creative.Bounds().Add(image.Pt(0, 0)), creative, image.ZP, draw.Over)
	draw.Draw(bgImg, barcode.Bounds().Add(image.Pt(20, 840)), barcode, image.ZP, draw.Over)

	outPath := barcodePath
	out, err := os.Create(outPath)
	if err != nil {
		return err
	}

	defer out.Close()

	err = png.Encode(out, bgImg)
	if err != nil {
		return err
	}

	return nil
}
