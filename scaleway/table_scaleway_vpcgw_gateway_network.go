package scaleway

import (
	"context"

	"github.com/scaleway/scaleway-sdk-go/api/vpcgw/v2"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

//// TABLE DEFINITION

func tableScalewayVPCGWGatewayNetwork(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:              "scaleway_vpcgw_gateway_network",
		Description:       "A gateway network attaches a Public Gateway to a private network.",
		GetMatrixItemFunc: BuildZoneList,
		List: &plugin.ListConfig{
			Hydrate:       listVPCGWGatewayNetworks,
			ParentHydrate: listVPCGWGateways,
			KeyColumns: []*plugin.KeyColumn{
				{
					Name:    "zone",
					Require: plugin.Optional,
				},
				{
					Name:    "gateway_id",
					Require: plugin.Optional,
				},
			},
		},
		Columns: []*plugin.Column{
			{
				Name:        "id",
				Description: "The unique identifier of the gateway network.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID"),
			},
			{
				Name:        "gateway_id",
				Description: "The ID of the gateway.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("GatewayID"),
			},
			{
				Name:        "private_network_id",
				Description: "The ID of the attached private network.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("PrivateNetworkID"),
			},
			{
				Name:        "status",
				Description: "The status of the attachment.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Status").Transform(transform.ToString),
			},
			{
				Name:        "mac_address",
				Description: "The MAC address of the gateway on the private network.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("MacAddress"),
			},
			{
				Name:        "masquerade_enabled",
				Description: "Indicates whether masquerading is enabled.",
				Type:        proto.ColumnType_BOOL,
				Transform:   transform.FromField("MasqueradeEnabled"),
			},
			{
				Name:        "push_default_route",
				Description: "Indicates whether the gateway pushes a default route to the private network.",
				Type:        proto.ColumnType_BOOL,
				Transform:   transform.FromField("PushDefaultRoute"),
			},
			{
				Name:        "ipam_ip_id",
				Description: "The ID of the IPAM IP address of the gateway on the private network.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("IpamIPID"),
			},
			{
				Name:        "srn",
				Description: "The Scaleway resource name of the gateway network.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Srn"),
			},
			{
				Name:        "created_at",
				Description: "The time when the gateway network was created.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("CreatedAt"),
			},
			{
				Name:        "updated_at",
				Description: "The time when the gateway network was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("UpdatedAt"),
			},

			// Scaleway standard columns
			{
				Name:        "region",
				Description: "The region where the gateway network resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Zone").Transform(lbRegionFromZone),
			},
			{
				Name:        "zone",
				Description: "The zone where the gateway network resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Zone").Transform(transform.ToString),
			},
			{
				Name:        "project_id",
				Description: "The ID of the project where the gateway network resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ProjectID"),
			},
			{
				Name:        "organization_id",
				Description: "The ID of the organization where the gateway network resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("OrganizationID"),
			},

			// Steampipe standard columns
			{
				Name:        "title",
				Description: "Title of the resource.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID"),
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

type vpcgwGatewayNetworkRow struct {
	vpcgw.GatewayNetwork
	ProjectID      string
	OrganizationID string
}

//// LIST FUNCTION

func listVPCGWGatewayNetworks(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	matrixLoc := d.EqualsQualString("zone")
	parent := h.Item.(*vpcgw.Gateway)

	if parent.Zone.String() != matrixLoc {
		return nil, nil
	}
	if id := d.EqualsQualString("gateway_id"); id != "" && id != parent.ID {
		return nil, nil
	}

	for _, item := range parent.GatewayNetworks {
		if item == nil {
			continue
		}
		d.StreamListItem(ctx, vpcgwGatewayNetworkRow{*item, parent.ProjectID, parent.OrganizationID})

		if d.RowsRemaining(ctx) == 0 {
			return nil, nil
		}
	}

	return nil, nil
}
