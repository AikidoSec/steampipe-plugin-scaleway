package scaleway

import (
	"context"

	secret "github.com/scaleway/scaleway-sdk-go/api/secret/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

//// TABLE DEFINITION

func tableScalewaySecret(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:              "scaleway_secret",
		Description:       "A secret stores sensitive data in Secret Manager. Only metadata is exposed, never version payloads.",
		GetMatrixItemFunc: BuildRegionList,
		List: &plugin.ListConfig{
			Hydrate: listSecrets,
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
				Description: "The unique identifier of the secret.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID"),
			},
			{
				Name:        "name",
				Description: "The name of the secret.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "path",
				Description: "The path of the secret.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "description",
				Description: "The description of the secret.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "status",
				Description: "The status of the secret.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Status").Transform(transform.ToString),
			},
			{
				Name:        "type",
				Description: "The type of the secret.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Type").Transform(transform.ToString),
			},
			{
				Name:        "version_count",
				Description: "The number of versions of the secret.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "managed",
				Description: "Indicates whether the secret is managed by another product.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "protected",
				Description: "Indicates whether the secret is protected against deletion.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "tags",
				Description: "A list of tags attached to the secret.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "used_by",
				Description: "The products using the secret.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "ephemeral_policy",
				Description: "The ephemeral policy of the secret.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "key_id",
				Description: "The ID of the key used to encrypt the secret.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("KeyID"),
			},
			{
				Name:        "srn",
				Description: "The Scaleway resource name of the secret.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Srn"),
			},
			{
				Name:        "deletion_requested_at",
				Description: "The time when the deletion of the secret was requested.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "created_at",
				Description: "The time when the secret was created.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "updated_at",
				Description: "The time when the secret was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
			},

			// Scaleway standard columns
			{
				Name:        "region",
				Description: "The region where the secret resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Region").Transform(transform.ToString),
			},
			{
				Name:        "project_id",
				Description: "The ID of the project where the secret resides.",
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

func listSecrets(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	region, err := scw.ParseRegion(d.EqualsQualString("region"))
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_secret.listSecrets", "region_parsing_error", err)
		return nil, err
	}

	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_secret.listSecrets", "connection_error", err)
		return nil, err
	}

	api := secret.NewAPI(client)

	req := &secret.ListSecretsRequest{
		Region:   region,
		Page:     scw.Int32Ptr(1),
		PageSize: scw.Uint32Ptr(listPageSize(d)),
	}

	var count int

	for {
		resp, err := api.ListSecrets(req)
		if err != nil {
			plugin.Logger(ctx).Error("scaleway_secret.listSecrets", "query_error", err)
			if isNotFoundOrUnsupportedZoneError(err) {
				return nil, nil
			}
			return nil, err
		}

		for _, item := range resp.Secrets {
			d.StreamListItem(ctx, item)
			count++

			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		if len(resp.Secrets) == 0 || uint64(count) >= uint64(resp.TotalCount) {
			break
		}
		req.Page = scw.Int32Ptr(*req.Page + 1)
	}

	return nil, nil
}
