package types

import (
	"testing"
)

func TestKeyer_String(t *testing.T) {
	t.Run("Keyer.String", func(t *testing.T) {
		tests := []struct {
			name string
			key  Keyer
			want string
		}{
			{
				name: "Test with non-empty string",
				key:  Keyer{Raw: "testKey"},
				want: "testKey",
			},
			{
				name: "Test with empty string",
				key:  Keyer{Raw: ""},
				want: "",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := tt.key.String(); got != tt.want {
					t.Errorf("Keyer.String() = %v, want %v", got, tt.want)
				}
			})
		}
	})
}
