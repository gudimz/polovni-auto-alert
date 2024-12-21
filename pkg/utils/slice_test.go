package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoveDuplicates(t *testing.T) {
	testCases := []struct {
		name     string
		input    any
		expected any
	}{
		{
			name:     "remove duplicates from int slice",
			input:    []int{1, 2, 3, 4, 5, 1, 2, 3, 4, 5},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "remove duplicates from string slice",
			input:    []string{"a", "b", "c", "d", "e", "a", "b", "c", "d", "e"},
			expected: []string{"a", "b", "c", "d", "e"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			switch input := tc.input.(type) {
			case []int:
				got := RemoveDuplicates(input)
				assert.Equal(t, tc.expected, got)
			case []string:
				got := RemoveDuplicates(input)
				assert.Equal(t, tc.expected, got)
			default:
				t.Errorf("unsupported type: %T", tc.input)
			}
		})
	}
}
