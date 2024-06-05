package ansi_test

import (
	"testing"

	"github.com/pynezz/pynezzentials/ansi"
)

// TestHexToRGB tests the HexToRGB function
func TestHexToRGB(t *testing.T) {
	tests := []struct {
		hex      string
		expected [3]int
		hasError bool
	}{
		{"#3498db", [3]int{52, 152, 219}, false},
		{"3498db", [3]int{52, 152, 219}, false},
		{"#000000", [3]int{0, 0, 0}, false},
		{"#ffffff", [3]int{255, 255, 255}, false},
		{"#gggggg", [3]int{0, 0, 0}, true}, // invalid hex
		{"#12345", [3]int{0, 0, 0}, true},  // invalid length
	}

	for _, test := range tests {
		r, g, b, err := ansi.HexToRGB(test.hex)
		if test.hasError {
			if err == nil {
				t.Errorf("Expected error for hex %s, but got none", test.hex)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for hex %s: %v", test.hex, err)
			}
			if r != test.expected[0] || g != test.expected[1] || b != test.expected[2] {
				t.Errorf("For hex %s, expected RGB (%d, %d, %d), but got (%d, %d, %d)", test.hex, test.expected[0], test.expected[1], test.expected[2], r, g, b)
			}
		}
	}
}

// TestHexColor256 tests the HexColor256 function
func TestHexColor256(t *testing.T) {
	tests := []struct {
		r, g, b  int
		msg      string
		expected string
	}{
		{52, 152, 219, "Hello, World!", "\033[38;5;68mHello, World!\033[0m"},
		{0, 0, 0, "Black", "\033[38;5;16mBlack\033[0m"},
		{255, 255, 255, "White", "\033[38;5;231mWhite\033[0m"},
	}

	for _, test := range tests {
		result := ansi.HexColor256(test.r, test.g, test.b, test.msg)
		if result != test.expected {
			t.Errorf("For RGB (%d, %d, %d) and msg %s, expected %s, but got %s", test.r, test.g, test.b, test.msg, test.expected, result)
		}
	}
}

// TestColor256 tests the Color256 function
func TestColor256(t *testing.T) {
	tests := []struct {
		color    ansi.U8Color
		msg      string
		expected string
	}{
		{"32", "Green Message", "\033[38;5;32mGreen Message\033[0m"},
		{"91", "Light Red Message", "\033[38;5;91mLight Red Message\033[0m"},
	}

	for _, test := range tests {
		result := ansi.U8Color("").Color256(test.color, test.msg)
		if result != test.expected {
			t.Errorf("For color %s and msg %s, expected %s, but got %s", test.color, test.msg, test.expected, result)
		}
	}
}
