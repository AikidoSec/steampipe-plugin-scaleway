package scaleway

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

type testStringEnum string

func TestHandleNilString(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  string
	}{
		{name: "versioning enabled", value: types.BucketVersioningStatusEnabled, want: "Enabled"},
		{name: "versioning suspended", value: types.BucketVersioningStatusSuspended, want: "Suspended"},
		{name: "versioning empty", value: types.BucketVersioningStatus(""), want: "false"},
		{name: "mfa delete enabled", value: types.MFADeleteStatusEnabled, want: "Enabled"},
		{name: "mfa delete empty", value: types.MFADeleteStatus(""), want: "false"},
		{name: "any string enum", value: testStringEnum("Custom"), want: "Custom"},
		{name: "plain string", value: "Enabled", want: "Enabled"},
		{name: "nil", value: nil, want: "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := handleNilString(context.Background(), &transform.TransformData{Value: tt.value})
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("handleNilString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandleNilStringToBool(t *testing.T) {
	tests := []struct {
		value interface{}
		want  bool
	}{
		{types.BucketVersioningStatusEnabled, true},
		{types.BucketVersioningStatusSuspended, false},
		{types.BucketVersioningStatus(""), false},
		{types.MFADeleteStatusEnabled, true},
	}

	for _, tt := range tests {
		str, err := handleNilString(context.Background(), &transform.TransformData{Value: tt.value})
		if err != nil {
			t.Fatal(err)
		}
		got, err := transform.ToBool(context.Background(), &transform.TransformData{Value: str})
		if err != nil {
			t.Fatal(err)
		}
		if got != tt.want {
			t.Errorf("bool(%v) = %v, want %v", tt.value, got, tt.want)
		}
	}
}

func TestOmitResultMetadata(t *testing.T) {
	str := func(s string) *string { return &s }

	t.Run("acl keeps owner and grants", func(t *testing.T) {
		acl := &s3.GetBucketAclOutput{
			Owner:  &types.Owner{ID: str("owner")},
			Grants: []types.Grant{{Permission: types.PermissionRead}},
		}
		got, err := omitResultMetadata(context.Background(), &transform.TransformData{Value: acl})
		if err != nil {
			t.Fatal(err)
		}
		fields := got.(map[string]interface{})
		if _, ok := fields["ResultMetadata"]; ok {
			t.Error("ResultMetadata must be removed")
		}
		if fields["Owner"] == nil || len(fields["Grants"].([]interface{})) != 1 {
			t.Errorf("other fields must be kept, got %v", fields)
		}
	})

	t.Run("website keeps index document", func(t *testing.T) {
		website := &s3.GetBucketWebsiteOutput{IndexDocument: &types.IndexDocument{Suffix: str("index.html")}}
		got, err := omitResultMetadata(context.Background(), &transform.TransformData{Value: website})
		if err != nil {
			t.Fatal(err)
		}
		if got.(map[string]interface{})["IndexDocument"] == nil {
			t.Errorf("IndexDocument must be kept, got %v", got)
		}
		if _, ok := got.(map[string]interface{})["ResultMetadata"]; ok {
			t.Error("ResultMetadata must be removed")
		}
	})

	t.Run("non-struct value passes through", func(t *testing.T) {
		got, err := omitResultMetadata(context.Background(), &transform.TransformData{Value: []string{"a", "b"}})
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, []interface{}{"a", "b"}) {
			t.Errorf("got %v", got)
		}
	})

	t.Run("nil stays nil", func(t *testing.T) {
		got, err := omitResultMetadata(context.Background(), &transform.TransformData{Value: nil})
		if err != nil || got != nil {
			t.Errorf("got %v, %v", got, err)
		}
	})
}

func TestFlattenObjectBucketTags(t *testing.T) {
	str := func(s string) *string { return &s }

	tests := []struct {
		name string
		in   []types.Tag
		want []string
	}{
		{name: "keys only", in: []types.Tag{{Key: str("env"), Value: str("prod")}, {Key: str("team")}}, want: []string{"env", "team"}},
		{name: "nil key skipped", in: []types.Tag{{Value: str("x")}, {Key: str("a")}}, want: []string{"a"}},
		{name: "empty", in: []types.Tag{}, want: nil},
		{name: "nil", in: nil, want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := flattenObjectBucketTags(tt.in); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
