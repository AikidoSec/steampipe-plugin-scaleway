package scaleway

import (
	"context"

	"github.com/scaleway/scaleway-sdk-go/api/mongodb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

//// TABLE DEFINITION

func tableScalewayMongoDBSnapshot(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:              "scaleway_mongodb_snapshot",
		Description:       "A Managed Database for MongoDB snapshot is a point-in-time copy of an instance.",
		GetMatrixItemFunc: BuildRegionList,
		List: &plugin.ListConfig{
			Hydrate: listMongoDBSnapshots,
			KeyColumns: []*plugin.KeyColumn{
				{
					Name:    "region",
					Require: plugin.Optional,
				},
				{
					Name:    "project_id",
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
				Name:        "instance_id",
				Description: "The ID of the instance the snapshot was taken from.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("InstanceID"),
			},
			{
				Name:        "instance_name",
				Description: "The name of the instance the snapshot was taken from.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "status",
				Description: "The status of the snapshot.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Status").Transform(transform.ToString),
			},
			{
				Name:        "size_bytes",
				Description: "The size of the snapshot, in bytes.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("SizeBytes").Transform(sizeToInt),
			},
			{
				Name:        "node_type",
				Description: "The node type of the instance at the time of the snapshot.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "volume_type",
				Description: "The volume type of the snapshot.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("VolumeType").Transform(transform.ToString),
			},
			{
				Name:        "srn",
				Description: "The Scaleway resource name of the snapshot.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Srn"),
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
				Description: "The ID of the project of the owning instance, null when the instance no longer exists.",
				Type:        proto.ColumnType_STRING,
				Hydrate:     getMongoDBSnapshotProjectID,
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

type mongodbSnapshotRow struct {
	mongodb.Snapshot
	InstanceID string
}

//// LIST FUNCTION

func listMongoDBSnapshots(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	region, err := scw.ParseRegion(d.EqualsQualString("region"))
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_mongodb_snapshot.listMongoDBSnapshots", "region_parsing_error", err)
		return nil, err
	}

	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_mongodb_snapshot.listMongoDBSnapshots", "connection_error", err)
		return nil, err
	}

	api := mongodb.NewAPI(client)

	req := &mongodb.ListSnapshotsRequest{
		Region:   region,
		Page:     scw.Int32Ptr(1),
		PageSize: scw.Uint32Ptr(listPageSize(d)),
	}
	if project := d.EqualsQualString("project_id"); project != "" {
		req.ProjectID = scw.StringPtr(project)
	}

	var count int

	for {
		resp, err := api.ListSnapshots(req)
		if err != nil {
			plugin.Logger(ctx).Error("scaleway_mongodb_snapshot.listMongoDBSnapshots", "query_error", err)
			if is403Error(err) || isNotFoundOrUnsupportedZoneError(err) {
				return nil, nil
			}
			return nil, err
		}

		for _, item := range resp.Snapshots {
			if item == nil {
				continue
			}
			var instanceID string
			if item.InstanceID != nil {
				instanceID = *item.InstanceID
			}
			d.StreamListItem(ctx, mongodbSnapshotRow{*item, instanceID})
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

//// HYDRATE FUNCTIONS

func getMongoDBSnapshotProjectID(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	row := h.Item.(mongodbSnapshotRow)
	region := row.Region
	if row.InstanceID == "" || region.String() != d.EqualsQualString("region") {
		return nil, nil
	}

	return cachedParentProjectID(ctx, d, "mongodb.instance.project."+region.String()+"."+row.InstanceID, func(client *scw.Client) (string, error) {
		instance, err := mongodb.NewAPI(client).GetInstance(&mongodb.GetInstanceRequest{Region: region, InstanceID: row.InstanceID})
		if err != nil {
			return "", err
		}
		return instance.ProjectID, nil
	})
}
