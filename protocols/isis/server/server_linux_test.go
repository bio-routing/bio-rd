package server

import(
	"testing"
)

func TestHtons(t *testing.T) {
	tests := []struct{
		name string
		input uint16
		expected uint16
	}{
		{
			name: "Test #1",
			input: 65024,
			expected: 0,
		},
	}

	for _, test := range tests {
		res := htons(test.input)
		if res != test.expected {
			t.Errorf("Test %q failed: Got: %v Expected: %v", test.name, res, test.expected)
		}
	}
}