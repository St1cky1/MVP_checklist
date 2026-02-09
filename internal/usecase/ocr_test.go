package usecase

import (
	"regexp"
	"strings"
	"testing"
)

func TestExtractSerialNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		found    bool
	}{
		{
			name: "Full sticker text",
			input: `
				|||||||||||||||||||
				XY202504160092
				物料 0050002126
				订单 MO056223
				ID : 2505000490
			`,
			expected: "XY202504160092",
			found:    true,
		},
		{
			name:     "Only serial",
			input:    "XY202504160092",
			expected: "XY202504160092",
			found:    true,
		},
		{
			name: "Serial with noise",
			input: "Noise AB123456789012 MoreNoise",
			expected: "AB123456789012",
			found:    true,
		},
		{
			name: "No valid serial",
			input: "0050002126 MO056223 ID:2505000490",
			expected: "",
			found:    false,
		},
	}

	re := regexp.MustCompile(`[A-Z]{2}\d{12}`)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := strings.Fields(tt.input)
			var result string
			var found bool
			for _, token := range tokens {
				if re.MatchString(token) {
					result = re.FindString(token)
					found = true
					break
				}
			}

			if found != tt.found {
				t.Errorf("Expected found=%v, got %v", tt.found, found)
			}
			if result != tt.expected {
				t.Errorf("Expected result=%s, got %s", tt.expected, result)
			}
		})
	}
}
