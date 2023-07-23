package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"io"
	"math"
	"os"
)

func byteCapacity(carrier io.Reader) (int, error) {
	config, _, err := image.DecodeConfig(carrier)
	if err != nil {
		return 0, err
	}
	imageSize := config.Height * config.Width
	capacity := int(math.Floor(float64(imageSize*3.0)/8.0)) - 1
	return capacity, nil
}

func hideBytes(carrier image.Image, message []byte) image.Image {
	var endOfTransmission byte = 0b0000_0100
	var bitMasks [8]byte = [8]byte{0b0000_0001, 0b0000_0010, 0b0000_0100, 0b0000_1000,
		0b0001_0000, 0b0010_0000, 0b0100_0000, 0b1000_0000}
	var rgbFactor uint32 = 257 // RGBA() returns images in the range [0, 65535] rather than [0, 255], but new colours need the second range.
	var byteIndex, bitIndex int = 0, 0

	message = append(message, endOfTransmission)
	messageComplete := false

	bounds := carrier.Bounds()
	newImage := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if !messageComplete {
				red, green, blue, _ := carrier.At(x, y).RGBA()
				var newRed, newGreen, newBlue uint8
				newGreen = uint8(green / rgbFactor)
				newBlue = uint8(blue / rgbFactor)

				if (message[byteIndex] & bitMasks[bitIndex]) > 0 {
					newRed = uint8(red/rgbFactor) | 0b0000_0001
				} else {
					newRed = uint8(red/rgbFactor) & 0b1111_1110
				}
				bitIndex++
				if bitIndex >= 8 {
					bitIndex = 0
					byteIndex++
					if byteIndex >= len(message) {
						messageComplete = true
						newImage.Set(x, y, color.RGBA{newRed, newGreen, newBlue, 255})
						continue
					}
				}

				if (message[byteIndex] & bitMasks[bitIndex]) > 0 {
					newGreen = uint8(green/rgbFactor) | 0b0000_0001
				} else {
					newGreen = uint8(green/rgbFactor) & 0b1111_1110
				}
				bitIndex++
				if bitIndex >= 8 {
					bitIndex = 0
					byteIndex++
					if byteIndex >= len(message) {
						messageComplete = true
						newImage.Set(x, y, color.RGBA{newRed, newGreen, newBlue, 255})
						continue
					}
				}

				if (message[byteIndex] & bitMasks[bitIndex]) > 0 {
					newBlue = uint8(blue/rgbFactor) | 0b0000_0001
				} else {
					newBlue = uint8(blue/rgbFactor) & 0b1111_1110
				}
				bitIndex++
				if bitIndex >= 8 {
					bitIndex = 0
					byteIndex++
					if byteIndex >= len(message) {
						messageComplete = true
					}
				}
				newImage.Set(x, y, color.RGBA{newRed, newGreen, newBlue, 255})
			} else {
				newImage.Set(x, y, carrier.At(x, y))
			}
		}
	}

	return newImage
}

func main() {
	path := "test.png"
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	existingImageFile, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer existingImageFile.Close()

	capacity, err := byteCapacity(existingImageFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Can hold %v bytes\n", capacity)

	messageSource := "test.txt"
	if len(os.Args) > 2 {
		messageSource = os.Args[2]
	}

	messageFile, err := os.ReadFile(messageSource)
	if err != nil {
		fmt.Println(err)
		return
	}

	if len(messageFile) > capacity {
		fmt.Printf("Source file will not fit in destination, %v bytes greater than %v byte capacity\n", len(messageFile), capacity)
		return
	}

	existingImageFile.Seek(0, 0)

	imageData, imageType, err := image.Decode(existingImageFile)
	if err != nil || imageData == nil {
		fmt.Println(err)
	}
	fmt.Println(imageType)

	codedImage := hideBytes(imageData, messageFile)

	newFile, err := os.Create("output.png")
	if err != nil {
		fmt.Println(err)
	}
	defer newFile.Close()

	err = png.Encode(newFile, codedImage)
	if err != nil {
		fmt.Println(err)
	}
}
