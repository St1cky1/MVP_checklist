package usecase

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/otiai10/gosseract/v2"
)

type OCRUseCase struct {
}

func NewOCRUseCase() *OCRUseCase {
	return &OCRUseCase{}
}

func (u *OCRUseCase) ProcessOCR(imageBytes []byte) (string, error) {
	// 1. Decode image
	img, _, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	// 2. Preprocessing for better OCR
	// - Convert to Grayscale
	// - Increase contrast
	// - Sharpen
	processedImg := imaging.Grayscale(img)
	processedImg = imaging.AdjustContrast(processedImg, 20)
	processedImg = imaging.Sharpen(processedImg, 0.5)

	// Encode back to bytes
	buf := new(bytes.Buffer)
	err = imaging.Encode(buf, processedImg, imaging.JPEG)
	if err != nil {
		return "", fmt.Errorf("failed to encode processed image: %w", err)
	}

	// 3. OCR with Tesseract
	client := gosseract.NewClient()
	defer client.Close()

	err = client.SetImageFromBytes(buf.Bytes())
	if err != nil {
		return "", fmt.Errorf("failed to set image for OCR: %w", err)
	}

	// Set whitelist for serial numbers if possible (optional, but good for accuracy)
	// client.SetVariable("tessedit_char_whitelist", "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-")

	text, err := client.Text()
	if err != nil {
		return "", fmt.Errorf("OCR failed: %w", err)
	}

	result := strings.TrimSpace(text)
	log.Printf("OCR Result: [%s]", result)

	return result, nil
}
