package utils

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// SanitizeForS3Metadata limpia texto para usarlo en headers de metadata de S3
// - Elimina acentos (á → a, ñ → n)
// - Convierte a ASCII
// - Reemplaza saltos de línea por espacios
// - Limita longitud si es necesario
func SanitizeForS3Metadata(text string, maxLength int) string {
	if text == "" {
		return ""
	}

	// 1. Reemplazar saltos de línea por espacios
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\r", " ")

	// 2. Eliminar acentos usando normalización Unicode
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, text)

	// 3. Reemplazar caracteres especiales comunes
	replacements := map[string]string{
		"ñ": "n",
		"Ñ": "N",
		"ü": "u",
		"Ü": "U",
	}
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}

	// 4. Eliminar múltiples espacios consecutivos
	result = strings.Join(strings.Fields(result), " ")

	// 5. Limitar longitud si es necesario
	if maxLength > 0 && len(result) > maxLength {
		result = result[:maxLength]
	}

	return result
}
