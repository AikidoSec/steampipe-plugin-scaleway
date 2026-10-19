package scaleway

import (
	"context"

	rdb "github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

//// TABLE DEFINITION

func tableScalewayRDBSnapshot(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:              "scaleway_rdb_snapshot",
		Description:       "An RDB snapshot is a point-in-time copy of a Database Instance volume.",
		GetMatrixItemFunc: BuildRegionList,
		List: &plugin.ListConfig{
			Hydrate: listRDBSnapshots,
			KeyColumns: []*plugin.KeyColumn{
				{
					Name:    "region",
					Require: plugin.Optional,
				},
			},
		},
		Columns: []*plugin.Column{
			{
				Name:        "id",
				Description: "The unique identifier of the snapshot.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID"),
			},
			{
				Name:        "name",
				Description: "The name of the snapshot.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "status",
				Description: "The status of the snapshot.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Status").Transform(transform.ToString),
			},
			{
				Name:        "instance_id",
				Description: "The ID of the Database Instance the snapshot was taken from.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("InstanceID"),
			},
			{
				Name:        "instance_name",
				Description: "The name of the Database Instance the snapshot was taken from.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "node_type",
				Description: "The node type of the Database Instance.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "volume_type",
				Description: "The volume type of the snapshot.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "size",
				Description: "The size of the snapshot in bytes.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("Size").Transform(sizeToInt),
			},
			{
				Name:        "expires_at",
				Description: "The time when the snapshot expires.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "created_at",
				Description: "The time when the snapshot was created.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "updated_at",
				Description: "The time when the snapshot was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
			},

			// Scaleway standard columns
			{
				Name:        "region",
				Description: "The region where the snapshot resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Region").Transform(transform.ToString),
			},
			{
				Name:        "project_id",
				Description: "The ID of the project of the owning Database Instance, null when the instance no longer exists.",
				Type:        proto.ColumnType_STRING,
				Hydrate:     getRDBInstanceProjectID,
				Transform:   transform.FromValue(),
			},

			// Steampipe standard columns
			{
				Name:        "title",
				Description: "Title of the resource.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Name"),
			},
			{
				Name:        "akas",
				Description: "Array of globally unique identifier strings (also known as) for the resource.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("ID").Transform(transform.EnsureStringArray),
			},
		},
	}
}

//// LIST FUNCTION

func listRDBSnapshots(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	region, err := scw.ParseRegion(d.EqualsQualString("region"))
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_rdb_snapshot.listRDBSnapshots", "region_parsing_error", err)
		return nil, err
	}

	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_rdb_snapshot.listRDBSnapshots", "connection_error", err)
		return nil, err
	}

	api := rdb.NewAPI(client)

	req := &rdb.ListSnapshotsRequest{
		Region:   region,
		Page:     scw.Int32Ptr(1),
		PageSize: scw.Uint32Ptr(listPageSize(d)),
	}

	var count int

	for {
		resp, err := api.ListSnapshots(req)
		if err != nil {
			plugin.Logger(ctx).Error("scaleway_rdb_snapshot.listRDBSnapshots", "query_error", err)
			if isNotFoundOrUnsupportedZoneError(err) {
				return nil, nil
			}
			return nil, err
		}

		for _, item := range resp.Snapshots {
			d.StreamListItem(ctx, item)
			count++

			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		if len(resp.Snapshots) == 0 || uint64(count) >= uint64(resp.TotalCount) {
			break
		}
		req.Page = scw.Int32Ptr(*req.Page + 1)
	}

	return nil, nil
}
