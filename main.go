package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
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
	/* We can store 3 bits in each pixel, so the image's total pixel count * 3 gives us a number of bytes.
	 * Divie by 8 for the number of bytes, which we floor, since we can't store a partial byte.
	 * 1 is subtracted from the total to make room for the EOT signal byte. */
	capacity := int(math.Floor(float64(imageSize*3.0)/8.0)) - 1
	return capacity, nil
}

func hideBytes(carrier image.Image, message []byte) image.Image {
	var endOfTransmission byte = 0b0000_0100
	var bitIndex int = 0

	// EOT is added to signal to the decoder that it's reached the end of the encoded message
	// NUL is added for the case where the message doesn't evenly divide by 3
	message = append(message, endOfTransmission, 0b0000_0000)
	messageLength := len(message) - 1

	// Find edges of the existing image and create the canvas we'll be editing
	bounds := carrier.Bounds()
	newImage := image.NewRGBA(bounds)
	draw.Draw(newImage, carrier.Bounds(), carrier, image.Point{}, draw.Over)

	// Iterate over image
	for y := bounds.Min.Y; y < bounds.Max.Y && (bitIndex>>3 < messageLength); y++ {
		for x := bounds.Min.X; x < bounds.Max.X && (bitIndex>>3 < messageLength); x++ {
			// For performance reasons, the RGB channels of a colour are left-shifted 8 times. First we undo that.
			red, green, blue, _ := carrier.At(x, y).RGBA()
			newRed, newGreen, newBlue := uint8(red>>8), uint8(green>>8), uint8(blue>>8)

			// Overwrite the least significant bit of each channel with a data bit
			newRed |= 0b0000_0001
			newRed &= (message[bitIndex>>3] >> (bitIndex & 0b0000_0111)) | 0b1111_1110
			bitIndex++
			newGreen |= 0b0000_0001
			newGreen &= (message[bitIndex>>3] >> (bitIndex & 0b0000_0111)) | 0b1111_1110
			bitIndex++
			newBlue |= 0b0000_0001
			newBlue &= (message[bitIndex>>3] >> (bitIndex & 0b0000_0111)) | 0b1111_1110
			bitIndex++

			newImage.Set(x, y, color.RGBA{newRed, newGreen, newBlue, 255})
		}
	}

	return newImage
}

func retrieveBytes(encodedImage image.Image) (message []byte) {
	var endOfTransmission byte = 0b0000_0100
	var bitIndex int = 0
	var tempByte byte

	bounds := encodedImage.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// For performance reasons, the RGB channels of a colour are left-shifted 8 times. First we undo that.
			red, green, blue, _ := encodedImage.At(x, y).RGBA()
			red |= red >> 8
			green |= green >> 8
			blue |= blue >> 8

			// Add the red channel's data bit to our storage byte
			tempByte |= ((byte(red) & 0b0000_0001) << bitIndex)
			bitIndex++
			if bitIndex >= 8 {
				// End of coded message
				if tempByte == endOfTransmission {
					return message
				}
				// Add completed byte to message, start reading next byte
				message = append(message, tempByte)
				tempByte = uint8(0)
				bitIndex = 0
			}

			// Add the green channel's data bit to our storage byte
			tempByte |= ((byte(green) & 0b0000_0001) << bitIndex)
			bitIndex++
			if bitIndex >= 8 {
				// End of coded message
				if tempByte == endOfTransmission {
					return message
				}
				// Add completed byte to message, start reading next byte
				message = append(message, tempByte)
				tempByte = uint8(0)
				bitIndex = 0
			}

			// Add the blue channel's data bit to our storage byte
			tempByte |= ((byte(blue) & 0b0000_0001) << bitIndex)
			bitIndex++
			if bitIndex >= 8 {
				// End of coded message
				if tempByte == endOfTransmission {
					return message
				}
				// Add completed byte to message, start reading next byte
				message = append(message, tempByte)
				tempByte = uint8(0)
				bitIndex = 0
			}
		}
	}

	return
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

	encodedImageFile, err := os.Open("output.png")
	if err != nil {
		fmt.Println(err)
		return
	}

	defer encodedImageFile.Close()

	encodedImage, _, err := image.Decode(encodedImageFile)
	if err != nil || encodedImage == nil {
		fmt.Println(err)
	}

	decodedBytes := retrieveBytes(encodedImage)

	os.WriteFile("output.txt", decodedBytes, 0666)
}
