package qr

import (
	"strings"

	"github.com/skip2/go-qrcode"
)

// GenerateASCII creates a small ASCII QR code for terminal.
// Uses block characters and is fast (<5ms for URL ~50 chars).
func GenerateASCII(url string) string {
	if url == "" {
		return ""
	}
	qr, err := qrcode.New(url, qrcode.Medium)
	if err != nil {
		return ""
	}
	bitmap := qr.Bitmap()
	var sb strings.Builder
	sb.WriteString(strings.Repeat("  ", len(bitmap[0])+2) + "\n")
	for _, row := range bitmap {
		sb.WriteString("  ")
		for _, col := range row {
			if col {
				sb.WriteString("██")
			} else {
				sb.WriteString("  ")
			}
		}
		sb.WriteString("  \n")
	}
	sb.WriteString(strings.Repeat("  ", len(bitmap[0])+2) + "\n")
	return sb.String()
}

// GeneratePNG returns QR code PNG bytes for web (256x256)
func GeneratePNG(url string) ([]byte, error) {
	if url == "" {
		return nil, nil
	}
	return qrcode.Encode(url, qrcode.Medium, 256)
}

func IsURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}
