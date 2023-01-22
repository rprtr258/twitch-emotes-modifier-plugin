package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReverseTimestamps(t *testing.T) {
	for _, test := range []struct {
		name string
		arg  []int
		want []int
	}{
		{
			name: "single frame",
			arg:  []int{0},
			want: []int{0},
		},
		{
			name: "simple",
			arg:  []int{1, 3, 6},
			want: []int{3, 5, 6},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			got := ReverseTimestamps(test.arg)
			assert.Equal(t, test.want, got)
		})
	}
}
