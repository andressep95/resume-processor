package utils

import (
	"encoding/json"
	"testing"
)

func TestIsValidDate(t *testing.T) {
	tests := []struct {
		name     string
		dateStr  string
		expected bool
	}{
		{"Empty string", "", true},
		{"Valid ISO date", "2024-01-15", true},
		{"Valid slash date", "15/01/2024", true},
		{"Valid year-month", "2024-01", true},
		{"Valid month-year", "01/2024", true},
		{"Valid year only", "2024", true},
		{"Valid month name", "Jan 2024", true},
		{"Invalid text", "No especificado", false},
		{"Invalid text 2", "Presente", false},
		{"Invalid text 3", "N/A", false},
		{"Invalid format", "2024/13/45", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidDate(tt.dateStr)
			if result != tt.expected {
				t.Errorf("IsValidDate(%q) = %v, expected %v", tt.dateStr, result, tt.expected)
			}
		})
	}
}

func TestSanitizeCertificationDates(t *testing.T) {
	jsonData := `{
		"certifications": [
			{
				"name": "AWS Certified",
				"dateObtained": "2024-01-15"
			},
			{
				"name": "Google Cloud",
				"dateObtained": "No especificado"
			},
			{
				"name": "Azure",
				"issueDate": "Presente",
				"expiryDate": "2025-12-31"
			}
		]
	}`

	var data map[string]interface{}
	json.Unmarshal([]byte(jsonData), &data)

	SanitizeCertificationDates(data)

	certs := data["certifications"].([]interface{})

	// Primera certificación: fecha válida, debe mantenerse
	cert1 := certs[0].(map[string]interface{})
	if cert1["dateObtained"] != "2024-01-15" {
		t.Errorf("Expected valid date to remain, got %v", cert1["dateObtained"])
	}

	// Segunda certificación: fecha inválida, debe limpiarse
	cert2 := certs[1].(map[string]interface{})
	if cert2["dateObtained"] != "" {
		t.Errorf("Expected invalid date to be empty, got %v", cert2["dateObtained"])
	}

	// Tercera certificación: issueDate inválida, expiryDate válida
	cert3 := certs[2].(map[string]interface{})
	if cert3["issueDate"] != "" {
		t.Errorf("Expected invalid issueDate to be empty, got %v", cert3["issueDate"])
	}
	if cert3["expiryDate"] != "2025-12-31" {
		t.Errorf("Expected valid expiryDate to remain, got %v", cert3["expiryDate"])
	}
}
