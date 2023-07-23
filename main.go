package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
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

	messageFile, err := os.Open(messageSource)
	if err != nil {
		fmt.Println(err)
		return
	}

	messageFileInfo, err := messageFile.Stat()
	if err != nil {
		fmt.Println(err)
		return
	}

	if messageFileInfo.Size() > int64(capacity) {
		fmt.Printf("Source file will not fit in destination, %v bytes greater than %v byte capacity\n", messageFileInfo.Size(), int64(capacity))
		return
	}

	existingImageFile.Seek(0, 0)

	imageData, imageType, err := image.Decode(existingImageFile)
	if err != nil || imageData == nil {
		fmt.Println(err)
	}
	fmt.Println(imageType)
}
