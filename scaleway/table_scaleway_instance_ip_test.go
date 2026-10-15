package scaleway

import (
	"context"
	"net"
	"testing"

	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"

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

func TestInstanceIPTitle(t *testing.T) {
	tests := []struct {
		name string
		item interface{}
		want interface{}
	}{
		{
			name: "IPv4 uses the address",
			item: &instance.IP{ID: "ip-id-1", Address: net.ParseIP("51.15.53.53")},
			want: "51.15.53.53",
		},
		{
			name: "IPv6 prefix row without an address falls back to the id",
			item: &instance.IP{ID: "ip-id-2", Type: instance.IPTypeRoutedIPv6},
			want: "ip-id-2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := instanceIPTitle(context.Background(), &transform.TransformData{HydrateItem: tt.item})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
