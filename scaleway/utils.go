package scaleway

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	gokittypes "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

// handleNilString maps an empty or nil value to "false". String-kinded enums (such as the S3 versioning
// statuses) are converted to string first because SafeString only recognises string and *string and
// would return "" for them.
func handleNilString(_ context.Context, d *transform.TransformData) (interface{}, error) {
	value := d.Value
	if rv := reflect.ValueOf(value); rv.IsValid() && rv.Kind() == reflect.String {
		value = rv.String()
	}

	str := gokittypes.SafeString(value)
	if str == "" {
		return "false", nil
	}
	return str, nil
}

// omitResultMetadata returns the transform value as plain JSON data without the ResultMetadata field
// that every aws-sdk-go-v2 operation output carries.
func omitResultMetadata(_ context.Context, d *transform.TransformData) (interface{}, error) {
	if d.Value == nil {
		return nil, nil
	}

	raw, err := json.Marshal(d.Value)
	if err != nil {
		return nil, err
	}

	var data interface{}
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}

	if fields, ok := data.(map[string]interface{}); ok {
		delete(fields, "ResultMetadata")
	}

	return data, nil
}

func flattenObjectBucketTags(tagsSet []types.Tag) []string {
	var tags []string

	for _, tagSet := range tagsSet {
		if tagSet.Key != nil {
			tags = append(tags, *tagSet.Key)
		}
	}

	return tags
}
