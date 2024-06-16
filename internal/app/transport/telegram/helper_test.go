package telegram

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_contains(t *testing.T) {
	testCases := []struct {
		name  string
		item  string
		slice []string
		want  bool
	}{
		{name: "item in slice", item: "hello", slice: []string{"hello", "world"}, want: true},
		{name: "item not in slice", item: "hello", slice: []string{"bye", "world"}, want: false},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			got := contains(tt.slice, tt.item)
			require.Equal(t, tt.want, got)
		})
	}
}
