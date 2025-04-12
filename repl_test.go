package main
import (
	"testing"
)
func TestCleanInput(t *testing.T){
	cases := []struct {
		input    string
		expected []string
	}{
		{
			input:    "  hello  world  ",
			expected: []string{"hello", "world"},
		},
		{
			input:    "hello",
			expected: []string{"hello"},
		},
		{
			input:    "   multiple   spaces   between   ",
			expected: []string{"multiple", "spaces", "between"},
		},
		{
			input:    "",
			expected: []string{},
		},
		{
			input:    "   ",
			expected: []string{},
		},
	}
	for _, c := range cases {
		actual := CleanInput(c.input)
		if len(actual) != len(c.expected) {
            t.Errorf("expected %d words, got %d for input: %q", len(c.expected), len(actual), c.input)
			continue
		}
		for i := range actual {
			word := actual[i]
			expectedWord := c.expected[i]
			if word != expectedWord {
                t.Errorf("expected word %q, got %q for input: %q", expectedWord, word, c.input)
			}
		}
	}
}