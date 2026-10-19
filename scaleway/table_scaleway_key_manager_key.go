package scaleway

import (
	"context"

	key_manager "github.com/scaleway/scaleway-sdk-go/api/key_manager/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

//// TABLE DEFINITION

func tableScalewayKeyManagerKey(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:              "scaleway_key_manager_key",
		Description:       "A Key Manager key is a cryptographic key managed by Scaleway. Only metadata is exposed, never key material.",
		GetMatrixItemFunc: BuildRegionList,
		List: &plugin.ListConfig{
			Hydrate: listKeyManagerKeys,
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
				Description: "The unique identifier of the key.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID"),
			},
			{
				Name:        "name",
				Description: "The name of the key.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "description",
				Description: "The description of the key.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "state",
				Description: "The state of the key.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("State").Transform(transform.ToString),
			},
			{
				Name:        "origin",
				Description: "The origin of the key material.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Origin").Transform(transform.ToString),
			},
			{
				Name:        "protection_level",
				Description: "The protection level of the key.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ProtectionLevel").Transform(transform.ToString),
			},
			{
				Name:        "usage",
				Description: "The usage (algorithm) of the key.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "rotation_policy",
				Description: "The rotation policy of the key.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "rotation_count",
				Description: "The number of times the key has been rotated.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "rotated_at",
				Description: "The time when the key was last rotated.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "protected",
				Description: "Indicates whether the key is protected against deletion.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "locked",
				Description: "Indicates whether the key is locked.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "tags",
				Description: "A list of tags attached to the key.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "srn",
				Description: "The Scaleway resource name of the key.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Srn"),
			},
			{
				Name:        "deletion_requested_at",
				Description: "The time when the deletion of the key was requested.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "created_at",
				Description: "The time when the key was created.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "updated_at",
				Description: "The time when the key was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
			},

			// Scaleway standard columns
			{
				Name:        "region",
				Description: "The region where the key resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Region").Transform(transform.ToString),
			},
			{
				Name:        "project_id",
				Description: "The ID of the project where the key resides.",
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

func listKeyManagerKeys(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	region, err := scw.ParseRegion(d.EqualsQualString("region"))
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_key_manager_key.listKeyManagerKeys", "region_parsing_error", err)
		return nil, err
	}

	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_key_manager_key.listKeyManagerKeys", "connection_error", err)
		return nil, err
	}

	api := key_manager.NewAPI(client)

	req := &key_manager.ListKeysRequest{
		Region:   region,
		Page:     scw.Int32Ptr(1),
		PageSize: scw.Uint32Ptr(listPageSize(d)),
	}

	var count int

	for {
		resp, err := api.ListKeys(req)
		if err != nil {
			plugin.Logger(ctx).Error("scaleway_key_manager_key.listKeyManagerKeys", "query_error", err)
			if isNotFoundOrUnsupportedZoneError(err) {
				return nil, nil
			}
			return nil, err
		}

		for _, item := range resp.Keys {
			d.StreamListItem(ctx, item)
			count++

			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		if len(resp.Keys) == 0 || uint64(count) >= uint64(resp.TotalCount) {
			break
		}
		req.Page = scw.Int32Ptr(*req.Page + 1)
	}

	return nil, nil
}
