package scaleway

import (
	"context"

	"github.com/scaleway/scaleway-sdk-go/api/flexibleip/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

//// TABLE DEFINITION

func tableScalewayFlexibleIP(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:              "scaleway_flexible_ip",
		Description:       "A flexible IP is a public IP address for Elastic Metal servers that can move between servers.",
		GetMatrixItemFunc: BuildZoneList,
		List: &plugin.ListConfig{
			Hydrate: listFlexibleIPs,
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
				Description: "The unique identifier of the flexible IP.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID"),
			},
			{
				Name:        "ip_address",
				Description: "The IP address, as a CIDR.",
				Type:        proto.ColumnType_CIDR,
				Transform:   transform.FromField("IPAddress").Transform(ipNetToString),
			},
			{
				Name:        "description",
				Description: "The description of the flexible IP.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Description"),
			},
			{
				Name:        "status",
				Description: "The status of the flexible IP.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Status").Transform(transform.ToString),
			},
			{
				Name:        "server_id",
				Description: "The ID of the server the flexible IP is attached to, if any.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ServerID"),
			},
			{
				Name:        "mac_address",
				Description: "The virtual MAC address attached to the flexible IP.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("MacAddress"),
			},
			{
				Name:        "reverse",
				Description: "The reverse DNS of the flexible IP.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Reverse"),
			},
			{
				Name:        "tags",
				Description: "A list of tags attached to the flexible IP.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Tags"),
			},
			{
				Name:        "srn",
				Description: "The Scaleway resource name of the flexible IP.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Srn"),
			},
			{
				Name:        "created_at",
				Description: "The time when the flexible IP was created.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("CreatedAt"),
			},
			{
				Name:        "updated_at",
				Description: "The time when the flexible IP was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("UpdatedAt"),
			},

			// Scaleway standard columns
			{
				Name:        "region",
				Description: "The region where the flexible IP resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Zone").Transform(lbRegionFromZone),
			},
			{
				Name:        "zone",
				Description: "The zone where the flexible IP resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Zone").Transform(transform.ToString),
			},
			{
				Name:        "project_id",
				Description: "The ID of the project where the flexible IP resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ProjectID"),
			},
			{
				Name:        "organization_id",
				Description: "The ID of the organization where the flexible IP resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("OrganizationID"),
			},

			// Steampipe standard columns
			{
				Name:        "title",
				Description: "Title of the resource.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("IPAddress").Transform(ipNetToString),
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

func listFlexibleIPs(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	zone, err := scw.ParseZone(d.EqualsQualString("zone"))
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_flexible_ip.listFlexibleIPs", "zone_parsing_error", err)
		return nil, err
	}

	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_flexible_ip.listFlexibleIPs", "connection_error", err)
		return nil, err
	}

	api := flexibleip.NewAPI(client)

	req := &flexibleip.ListFlexibleIPsRequest{
		Zone:     zone,
		Page:     scw.Int32Ptr(1),
		PageSize: scw.Uint32Ptr(listPageSize(d)),
	}

	var count int

	for {
		resp, err := api.ListFlexibleIPs(req)
		if err != nil {
			plugin.Logger(ctx).Error("scaleway_flexible_ip.listFlexibleIPs", "query_error", err)
			if isNotFoundOrUnsupportedZoneError(err) {
				return nil, nil
			}
			return nil, err
		}

		for _, item := range resp.FlexibleIPs {
			d.StreamListItem(ctx, item)
			count++

			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		if len(resp.FlexibleIPs) == 0 || uint64(count) >= uint64(resp.TotalCount) {
			break
		}
		req.Page = scw.Int32Ptr(*req.Page + 1)
	}

	return nil, nil
}
