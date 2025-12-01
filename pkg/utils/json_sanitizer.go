package utils

import (
	"encoding/json"
	"regexp"
	"time"
)

// Patrones de fecha comunes
var datePatterns = []*regexp.Regexp{
	regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`),                    // 2024-01-15
	regexp.MustCompile(`^\d{2}/\d{2}/\d{4}$`),                    // 15/01/2024
	regexp.MustCompile(`^\d{4}/\d{2}/\d{2}$`),                    // 2024/01/15
	regexp.MustCompile(`^\d{4}-\d{2}$`),                          // 2024-01
	regexp.MustCompile(`^\d{2}/\d{4}$`),                          // 01/2024
	regexp.MustCompile(`^(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)\s+\d{4}$`), // Jan 2024
	regexp.MustCompile(`^\d{4}$`),                                // 2024
}

// IsValidDate verifica si un string es una fecha válida
func IsValidDate(dateStr string) bool {
	if dateStr == "" {
		return true // Vacío es válido
	}

	// Verificar patrones comunes primero
	for _, pattern := range datePatterns {
		if pattern.MatchString(dateStr) {
			// Validar que los números sean razonables
			if !isReasonableDate(dateStr) {
				return false
			}
			return true
		}
	}

	// Intentar parsear como fecha ISO
	formats := []string{
		time.RFC3339,
		"2006-01-02",
		"02/01/2006",
		"2006/01/02",
		"2006-01",
		"01/2006",
	}

	for _, format := range formats {
		if _, err := time.Parse(format, dateStr); err == nil {
			return true
		}
	}

	return false
}

// isReasonableDate verifica que los valores numéricos sean razonables
func isReasonableDate(dateStr string) bool {
	// Extraer números del string
	digitsPattern := regexp.MustCompile(`\d+`)
	matches := digitsPattern.FindAllString(dateStr, -1)

	for _, match := range matches {
		if len(match) == 4 { // Año
			year := 0
			for _, c := range match {
				year = year*10 + int(c-'0')
			}
			if year < 1900 || year > 2100 {
				return false
			}
		} else if len(match) == 2 { // Mes o día
			val := 0
			for _, c := range match {
				val = val*10 + int(c-'0')
			}
			if val < 1 || val > 31 {
				return false
			}
		}
	}

	return true
}

// ValidateCertificationDateFormat valida que la fecha esté en formato MM YYYY
func ValidateCertificationDateFormat(dateStr string) bool {
	if dateStr == "" {
		return true // Vacío es válido
	}

	// Patrón específico: MM YYYY (ej: "01 2024", "12 2023")
	pattern := regexp.MustCompile(`^(0[1-9]|1[0-2])\s+\d{4}$`)
	return pattern.MatchString(dateStr)
}

// SanitizeCertificationDates limpia las fechas en certificaciones del JSON
func SanitizeCertificationDates(structuredData map[string]interface{}) {
	// Buscar certificaciones en el JSON
	certifications, ok := structuredData["certifications"].([]interface{})
	if !ok {
		return
	}

	for _, cert := range certifications {
		certMap, ok := cert.(map[string]interface{})
		if !ok {
			continue
		}

		// Lista de campos de fecha a validar
		dateFields := []string{"issueDate", "expiryDate", "date", "dateObtained"}

		for _, field := range dateFields {
			if dateValue, exists := certMap[field]; exists {
				if dateStr, ok := dateValue.(string); ok {
					// Validación específica: solo MM YYYY es válido
					if !ValidateCertificationDateFormat(dateStr) {
						certMap[field] = ""
					}
				}
			}
		}
	}
}

// SanitizeStructuredData sanitiza todo el JSON recibido de AWS
func SanitizeStructuredData(data interface{}) (map[string]interface{}, error) {
	// Convertir a map
	var structuredData map[string]interface{}

	switch v := data.(type) {
	case map[string]interface{}:
		structuredData = v
	case []byte:
		if err := json.Unmarshal(v, &structuredData); err != nil {
			return nil, err
		}
	case string:
		if err := json.Unmarshal([]byte(v), &structuredData); err != nil {
			return nil, err
		}
	default:
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(jsonBytes, &structuredData); err != nil {
			return nil, err
		}
	}

	// Sanitizar certificaciones
	SanitizeCertificationDates(structuredData)

	return structuredData, nil
}
