package scaleway

import (
	"context"

	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

const maxListPageSize = 100

// listPageSize returns the page size for a list call: the API maximum, lowered to the
// query's row limit when that is smaller.
func listPageSize(d *plugin.QueryData) uint32 {
	size := int64(maxListPageSize)
	if limit := d.QueryContext.Limit; limit != nil && *limit > 0 && *limit < size {
		size = *limit
	}
	return uint32(size)
}

// sizeToInt renders a scw.Size or *scw.Size as an int64, or nil when unset.
func sizeToInt(_ context.Context, d *transform.TransformData) (interface{}, error) {
	switch v := d.Value.(type) {
	case scw.Size:
		return int64(v), nil
	case *scw.Size:
		if v == nil {
			return nil, nil
		}
		return int64(*v), nil
	}
	return nil, nil
}

// getRDBInstanceProjectID returns the project ID of the RDB instance that owns a snapshot or backup row,
// or nil when the instance is gone, not readable, or the row has no instance_id.
func getRDBInstanceProjectID(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	var instanceID string
	var region scw.Region
	switch item := h.Item.(type) {
	case *rdb.Snapshot:
		instanceID, region = item.InstanceID, item.Region
	case *rdb.DatabaseBackup:
		instanceID, region = item.InstanceID, item.Region
	default:
		return nil, nil
	}

	if instanceID == "" || region.String() != d.EqualsQualString("region") {
		return nil, nil
	}

	cacheKey := "rdb.instance.project." + region.String() + "." + instanceID
	if cached, ok := d.ConnectionManager.Cache.Get(cacheKey); ok {
		projectID, _ := cached.(string)
		if projectID == "" {
			return nil, nil
		}
		return projectID, nil
	}

	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway.getRDBInstanceProjectID", "connection_error", err)
		return nil, err
	}

	instance, err := rdb.NewAPI(client).GetInstance(&rdb.GetInstanceRequest{Region: region, InstanceID: instanceID})
	if err != nil {
		if is403Error(err) || isNotFoundOrUnsupportedZoneError(err) {
			d.ConnectionManager.Cache.Set(cacheKey, "")
			return nil, nil
		}
		plugin.Logger(ctx).Error("scaleway.getRDBInstanceProjectID", "query_error", err)
		return nil, err
	}

	d.ConnectionManager.Cache.Set(cacheKey, instance.ProjectID)
	if instance.ProjectID == "" {
		return nil, nil
	}
	return instance.ProjectID, nil
}
