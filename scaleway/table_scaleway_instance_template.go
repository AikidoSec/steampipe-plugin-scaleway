package scaleway

import (
	"context"

	instancev2 "github.com/scaleway/scaleway-sdk-go/api/instance/v2alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

//// TABLE DEFINITION

func tableScalewayInstanceTemplate(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:              "scaleway_instance_template",
		Description:       "An Instance template describes the configuration used to create servers.",
		GetMatrixItemFunc: BuildZoneList,
		List: &plugin.ListConfig{
			Hydrate: listInstanceTemplates,
			KeyColumns: []*plugin.KeyColumn{
				{
					Name:    "zone",
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
				Description: "The unique identifier of the template.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID"),
			},
			{
				Name:        "name",
				Description: "The name of the template.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Name"),
			},
			{
				Name:        "server_type",
				Description: "The commercial type of the servers created from the template.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ServerType"),
			},
			{
				Name:        "server_tags",
				Description: "The tags applied to servers created from the template.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("ServerTags"),
			},
			{
				Name:        "security_group_id",
				Description: "The ID of the security group of the template.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("SecurityGroupID"),
			},
			{
				Name:        "placement_group_id",
				Description: "The ID of the placement group of the template.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("PlacementGroupID"),
			},
			{
				Name:        "public_ip_v4_count",
				Description: "The number of public IPv4 addresses.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("PublicIPV4Count"),
			},
			{
				Name:        "public_ip_v6_count",
				Description: "The number of public IPv6 addresses.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("PublicIPV6Count"),
			},
			{
				Name:        "filesystem_ids",
				Description: "The IDs of the filesystems attached to the template.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("FilesystemIDs"),
			},
			{
				Name:        "tags",
				Description: "A list of tags attached to the template.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Tags"),
			},
			{
				Name:        "srn",
				Description: "The Scaleway resource name of the template.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Srn"),
			},
			{
				Name:        "created_at",
				Description: "The time when the template was created.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("CreatedAt"),
			},
			{
				Name:        "updated_at",
				Description: "The time when the template was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("UpdatedAt"),
			},

			// Scaleway standard columns
			{
				Name:        "region",
				Description: "The region where the template resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Zone").Transform(lbRegionFromZone),
			},
			{
				Name:        "zone",
				Description: "The zone where the template resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Zone").Transform(transform.ToString),
			},
			{
				Name:        "project_id",
				Description: "The ID of the project where the template resides.",
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

func listInstanceTemplates(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	zone, err := scw.ParseZone(d.EqualsQualString("zone"))
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_instance_template.listInstanceTemplates", "zone_parsing_error", err)
		return nil, err
	}

	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_instance_template.listInstanceTemplates", "connection_error", err)
		return nil, err
	}

	api := instancev2.NewAPI(client)

	req := &instancev2.ListTemplatesRequest{
		Zone:     zone,
		PageSize: scw.Uint32Ptr(listPageSize(d)),
	}
	if project := d.EqualsQualString("project_id"); project != "" {
		req.ProjectID = project
	}

	for {
		resp, err := api.ListTemplates(req)
		if err != nil {
			plugin.Logger(ctx).Error("scaleway_instance_template.listInstanceTemplates", "query_error", err)
			if isNotFoundOrUnsupportedZoneError(err) {
				return nil, nil
			}
			return nil, err
		}

		for _, item := range resp.Templates {
			d.StreamListItem(ctx, item)

			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		if resp.NextPageToken == nil || *resp.NextPageToken == "" {
			break
		}
		req.PageToken = resp.NextPageToken
	}

	return nil, nil
}
