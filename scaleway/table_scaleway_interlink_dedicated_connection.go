package scaleway

import (
	"context"

	"github.com/scaleway/scaleway-sdk-go/api/interlink/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

//// TABLE DEFINITION

func tableScalewayInterlinkDedicatedConnection(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:              "scaleway_interlink_dedicated_connection",
		Description:       "An InterLink dedicated connection is a physical connection between a customer and a Scaleway point of presence.",
		GetMatrixItemFunc: BuildRegionList,
		List: &plugin.ListConfig{
			Hydrate: listInterlinkDedicatedConnections,
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
				Description: "The unique identifier of the dedicated connection.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID"),
			},
			{
				Name:        "name",
				Description: "The name of the dedicated connection.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Name"),
			},
			{
				Name:        "status",
				Description: "The status of the connection.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Status").Transform(transform.ToString),
			},
			{
				Name:        "pop_id",
				Description: "The ID of the point of presence hosting the connection.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("PopID"),
			},
			{
				Name:        "bandwidth_mbps",
				Description: "The bandwidth of the connection, in Mbps.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("BandwidthMbps"),
			},
			{
				Name:        "available_link_bandwidths",
				Description: "The link bandwidths still available on the connection, in Mbps.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("AvailableLinkBandwidths"),
			},
			{
				Name:        "demarcation_info",
				Description: "The demarcation information of the connection.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("DemarcationInfo"),
			},
			{
				Name:        "vlan_range",
				Description: "The VLAN range available for links on the connection.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("VlanRange"),
			},
			{
				Name:        "tags",
				Description: "A list of tags attached to the dedicated connection.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Tags"),
			},
			{
				Name:        "srn",
				Description: "The Scaleway resource name of the dedicated connection.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Srn"),
			},
			{
				Name:        "created_at",
				Description: "The time when the dedicated connection was created.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("CreatedAt"),
			},
			{
				Name:        "updated_at",
				Description: "The time when the dedicated connection was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("UpdatedAt"),
			},

			// Scaleway standard columns
			{
				Name:        "region",
				Description: "The region where the dedicated connection resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Region").Transform(transform.ToString),
			},
			{
				Name:        "project_id",
				Description: "The ID of the project where the dedicated connection resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ProjectID"),
			},
			{
				Name:        "organization_id",
				Description: "The ID of the organization where the dedicated connection resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("OrganizationID"),
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

func listInterlinkDedicatedConnections(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	region, err := scw.ParseRegion(d.EqualsQualString("region"))
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_interlink_dedicated_connection.listInterlinkDedicatedConnections", "region_parsing_error", err)
		return nil, err
	}

	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_interlink_dedicated_connection.listInterlinkDedicatedConnections", "connection_error", err)
		return nil, err
	}

	api := interlink.NewAPI(client)

	req := &interlink.ListDedicatedConnectionsRequest{
		Region:   region,
		Page:     scw.Int32Ptr(1),
		PageSize: scw.Uint32Ptr(listPageSize(d)),
	}

	var count int

	for {
		resp, err := api.ListDedicatedConnections(req)
		if err != nil {
			plugin.Logger(ctx).Error("scaleway_interlink_dedicated_connection.listInterlinkDedicatedConnections", "query_error", err)
			if isNotFoundOrUnsupportedZoneError(err) {
				return nil, nil
			}
			return nil, err
		}

		for _, item := range resp.Connections {
			d.StreamListItem(ctx, item)
			count++

			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		if len(resp.Connections) == 0 || uint64(count) >= uint64(resp.TotalCount) {
			break
		}
		req.Page = scw.Int32Ptr(*req.Page + 1)
	}

	return nil, nil
}
