package utils

import (
	"testing"
)

func TestGenerateShortCode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int // Expected length of the short code
	}{
		{"Standard URL", "https://example.com", 6},
		{"Another URL", "https://google.com", 6},
		{"Subdomain URL", "https://sub.example.com", 6},
		{"Empty String", "", 6},
		{"Long URL", "https://example.com/" + string(make([]byte, 1000)), 6},
	}

	seen := make(map[string]bool)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := GenerateShortCode(test.input)

			// Check length
			if len(result) != test.expected {
				t.Errorf("Expected length %d, but got %d for input %s", test.expected, len(result), test.input)
			}

			// Check uniqueness
			if seen[result] {
				t.Errorf("Duplicate short code generated: %s", result)
			}
			seen[result] = true

			// Check consistency
			repeatedResult := GenerateShortCode(test.input)
			if result != repeatedResult {
				t.Errorf("Expected consistent output, but got different results: %s and %s", result, repeatedResult)
			}
		})
	}
}
