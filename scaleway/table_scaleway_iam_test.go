package scaleway

import (
	"context"
	"encoding/json"
	"testing"

	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func TestIamRulesToJSON(t *testing.T) {
	empty := []string{}
	names := []string{"IAMReadOnly"}
	projects := []string{"p1"}

	rules := []*iam.Rule{
		{ID: "nil"},
		{ID: "empty", PermissionSetNames: &empty, ProjectIDs: &empty},
		{ID: "full", PermissionSetNames: &names, ProjectIDs: &projects},
	}

	got, err := iamRulesToJSON(context.Background(), &transform.TransformData{Value: rules})
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}

	var decoded []map[string]interface{}
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded) != 3 {
		t.Fatalf("got %d rules, want 3", len(decoded))
	}
	if v, ok := decoded[0]["permission_set_names"]; !ok || v != nil {
		t.Errorf("nil pointer: got %v, want null", v)
	}
	if v, ok := decoded[0]["project_ids"]; !ok || v != nil {
		t.Errorf("nil pointer project_ids: got %v, want null", v)
	}
	if v, ok := decoded[1]["permission_set_names"].([]interface{}); !ok || len(v) != 0 {
		t.Errorf("empty slice: got %v, want []", decoded[1]["permission_set_names"])
	}
	if v, ok := decoded[2]["permission_set_names"].([]interface{}); !ok || len(v) != 1 || v[0] != "IAMReadOnly" {
		t.Errorf("populated: got %v", decoded[2]["permission_set_names"])
	}
	if v, ok := decoded[2]["project_ids"].([]interface{}); !ok || len(v) != 1 || v[0] != "p1" {
		t.Errorf("populated project_ids: got %v", decoded[2]["project_ids"])
	}
}

func TestIamRulesToJSONNonRuleValue(t *testing.T) {
	got, err := iamRulesToJSON(context.Background(), &transform.TransformData{Value: nil})
	if err != nil || got != nil {
		t.Fatalf("got %v, %v; want nil, nil", got, err)
	}
}

func TestDurationToSeconds(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  interface{}
	}{
		{"nil pointer", (*scw.Duration)(nil), nil},
		{"zero", &scw.Duration{}, int64(0)},
		{"populated", &scw.Duration{Seconds: 3600, Nanos: 5}, int64(3600)},
		{"wrong type", "x", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := durationToSeconds(context.Background(), &transform.TransformData{Value: tt.value})
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
