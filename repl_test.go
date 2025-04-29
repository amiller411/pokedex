// main_test.go
package main

import (
	"fmt"
	"testing"
)

func TestCleanInput(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{
			input:    " hello world",
			expected: []string{"hello", "world"},
		},
		{
			input:    "heya drew",
			expected: []string{"heya", "drew"},
		},
	}
	for _, c := range cases {
		actual := cleanInput(c.input)
		fmt.Println("INPUT:", c.input)
		fmt.Println("EXPECTED: ", c.expected)
		fmt.Println("ACTUAL: ", actual)
		actualLen := len(actual)
		expectedLen := len(c.expected)
		if actualLen != expectedLen {
			t.Errorf("Length mismatch for input '%s': got %d, want %d", c.input, actualLen, expectedLen)
		}

		for i := range actual {
			if i < expectedLen {
				if actual[i] != c.expected[i] {
					t.Errorf("Word mismatch at position %d for input '%s': got '%s', want '%s'",
						i, c.input, actual[i], c.expected[i])
				}
			}
		}
	}
}
