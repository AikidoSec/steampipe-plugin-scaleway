package scaleway

import (
	"context"

	"github.com/scaleway/scaleway-sdk-go/api/block/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

//// TABLE DEFINITION

func tableScalewayBlockSnapshot(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:              "scaleway_block_snapshot",
		Description:       "A block snapshot is a point-in-time copy of a block volume.",
		GetMatrixItemFunc: BuildZoneList,
		List: &plugin.ListConfig{
			Hydrate: listBlockSnapshots,
			KeyColumns: []*plugin.KeyColumn{
				{
					Name:    "zone",
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
				Name:        "class",
				Description: "The storage class of the snapshot.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Class").Transform(transform.ToString),
			},
			{
				Name:        "size",
				Description: "The size of the snapshot in bytes.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("Size").Transform(sizeToInt),
			},
			{
				Name:        "public",
				Description: "Indicates whether the snapshot is public.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "parent_volume",
				Description: "The volume the snapshot was taken from.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "references",
				Description: "The resources referencing the snapshot.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "kms_key_id",
				Description: "The ID of the KMS key encrypting the snapshot.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("KmsKeyID"),
			},
			{
				Name:        "tags",
				Description: "A list of tags attached to the snapshot.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "srn",
				Description: "The Scaleway resource name of the snapshot.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Srn"),
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
				Transform:   transform.FromField("Zone").Transform(lbRegionFromZone),
			},
			{
				Name:        "zone",
				Description: "The zone where the snapshot resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Zone").Transform(transform.ToString),
			},
			{
				Name:        "project_id",
				Description: "The ID of the project where the snapshot resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ProjectID"),
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

func listBlockSnapshots(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	zone, err := scw.ParseZone(d.EqualsQualString("zone"))
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_block_snapshot.listBlockSnapshots", "zone_parsing_error", err)
		return nil, err
	}

	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_block_snapshot.listBlockSnapshots", "connection_error", err)
		return nil, err
	}

	api := block.NewAPI(client)

	req := &block.ListSnapshotsRequest{
		Zone:     zone,
		Page:     scw.Int32Ptr(1),
		PageSize: scw.Uint32Ptr(listPageSize(d)),
	}

	var count int

	for {
		resp, err := api.ListSnapshots(req)
		if err != nil {
			plugin.Logger(ctx).Error("scaleway_block_snapshot.listBlockSnapshots", "query_error", err)
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
