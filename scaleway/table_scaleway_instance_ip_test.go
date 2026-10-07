package scaleway

import (
	"context"
	"net"
	"testing"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func TestSafeIPAddrToString(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  interface{}
	}{
		{
			name:  "populated IPv4 address",
			value: net.ParseIP("51.15.53.53"),
			want:  "51.15.53.53",
		},
		{
			name:  "empty net.IP",
			value: net.IP{},
			want:  nil,
		},
		{
			name:  "nil net.IP",
			value: net.IP(nil),
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := safeIPAddrToString(context.Background(), &transform.TransformData{Value: tt.value})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
