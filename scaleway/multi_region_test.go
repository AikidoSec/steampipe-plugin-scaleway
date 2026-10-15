package scaleway

import (
	"slices"
	"testing"

	"github.com/scaleway/scaleway-sdk-go/scw"
)

func TestZonesOfRegion(t *testing.T) {
	tests := []struct {
		region scw.Region
		want   []scw.Zone
	}{
		{"it-mil", []scw.Zone{"it-mil-1", "it-mil-2", "it-mil-3"}},
		{"fr-par", []scw.Zone{"fr-par-1", "fr-par-2", "fr-par-3"}},
		{"nl-ams", []scw.Zone{"nl-ams-1", "nl-ams-2", "nl-ams-3"}},
		{"pl-waw", []scw.Zone{"pl-waw-1", "pl-waw-2", "pl-waw-3"}},
		{"xx-unknown", nil},
		{"", nil},
	}

	for _, tt := range tests {
		t.Run(string(tt.region), func(t *testing.T) {
			if got := zonesOfRegion(tt.region); !slices.Equal(got, tt.want) {
				t.Errorf("zonesOfRegion(%q) = %v, want %v", tt.region, got, tt.want)
			}
		})
	}
}
