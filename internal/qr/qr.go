package qr

import (
	"strings"

	"github.com/skip2/go-qrcode"
)

// GenerateASCII creates a compact ASCII QR code for terminal (<15 linhas, ~30 colunas).
// Usa half-blocks (▀▄█ ) para representar 2 módulos por linha, sem borda extra.
func GenerateASCII(url string) string {
	if url == "" {
		return ""
	}
	// Low recovery deixa QR menor para URLs curtas (~25x25)
	qr, err := qrcode.New(url, qrcode.Low)
	if err != nil {
		return ""
	}
	bitmap := qr.Bitmap()
	// bitmap é quadrado, sem quiet zone extra - qrcode já inclui 4 módulos de borda
	var sb strings.Builder
	// Renderiza 2 linhas do QR por 1 linha do terminal usando half-blocks
	for y := 0; y < len(bitmap); y += 2 {
		for _, row := range [][]bool{bitmap[y]} {
			_ = row
		}
		for x := 0; x < len(bitmap[0]); x++ {
			top := bitmap[y][x]
			bottom := false
			if y+1 < len(bitmap) {
				bottom = bitmap[y+1][x]
			}
			switch {
			case top && bottom:
				sb.WriteString("█")
			case top && !bottom:
				sb.WriteString("▀")
			case !top && bottom:
				sb.WriteString("▄")
			default:
				sb.WriteString(" ")
			}
		}
		sb.WriteString("\n")
	}
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
