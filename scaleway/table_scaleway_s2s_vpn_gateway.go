package scaleway

import (
	"context"

	"github.com/scaleway/scaleway-sdk-go/api/s2s_vpn/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

//// TABLE DEFINITION

func tableScalewayS2SVPNGateway(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:              "scaleway_s2s_vpn_gateway",
		Description:       "A site-to-site VPN gateway is the Scaleway side of a VPN connection.",
		GetMatrixItemFunc: BuildRegionList,
		List: &plugin.ListConfig{
			Hydrate: listS2SVPNGateways,
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
				Description: "The unique identifier of the VPN gateway.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID"),
			},
			{
				Name:        "name",
				Description: "The name of the VPN gateway.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Name"),
			},
			{
				Name:        "status",
				Description: "The status of the gateway.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Status").Transform(transform.ToString),
			},
			{
				Name:        "gateway_type",
				Description: "The type of the gateway.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("GatewayType"),
			},
			{
				Name:        "private_network_id",
				Description: "The ID of the private network the gateway is attached to.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("PrivateNetworkID"),
			},
			{
				Name:        "ipam_private_ipv4_id",
				Description: "The ID of the IPAM private IPv4 address of the gateway.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("IpamPrivateIPv4ID"),
			},
			{
				Name:        "ipam_private_ipv6_id",
				Description: "The ID of the IPAM private IPv6 address of the gateway.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("IpamPrivateIPv6ID"),
			},
			{
				Name:        "asn",
				Description: "The BGP autonomous system number of the gateway.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("Asn"),
			},
			{
				Name:        "connection_ids",
				Description: "The IDs of the connections using the gateway.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("ConnectionIDs"),
			},
			{
				Name:        "zone",
				Description: "The zone where the gateway resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Zone").Transform(transform.ToString),
			},
			{
				Name:        "tags",
				Description: "A list of tags attached to the VPN gateway.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Tags"),
			},
			{
				Name:        "srn",
				Description: "The Scaleway resource name of the VPN gateway.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Srn"),
			},
			{
				Name:        "created_at",
				Description: "The time when the VPN gateway was created.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("CreatedAt"),
			},
			{
				Name:        "updated_at",
				Description: "The time when the VPN gateway was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("UpdatedAt"),
			},

			// Scaleway standard columns
			{
				Name:        "region",
				Description: "The region where the VPN gateway resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Region").Transform(transform.ToString),
			},
			{
				Name:        "project_id",
				Description: "The ID of the project where the VPN gateway resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ProjectID"),
			},
			{
				Name:        "organization_id",
				Description: "The ID of the organization where the VPN gateway resides.",
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

func listS2SVPNGateways(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	region, err := scw.ParseRegion(d.EqualsQualString("region"))
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_s2s_vpn_gateway.listS2SVPNGateways", "region_parsing_error", err)
		return nil, err
	}

	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_s2s_vpn_gateway.listS2SVPNGateways", "connection_error", err)
		return nil, err
	}

	api := s2s_vpn.NewAPI(client)

	req := &s2s_vpn.ListVpnGatewaysRequest{
		Region:   region,
		Page:     scw.Int32Ptr(1),
		PageSize: scw.Uint32Ptr(listPageSize(d)),
	}

	var count int

	for {
		resp, err := api.ListVpnGateways(req)
		if err != nil {
			plugin.Logger(ctx).Error("scaleway_s2s_vpn_gateway.listS2SVPNGateways", "query_error", err)
			if isNotFoundOrUnsupportedZoneError(err) {
				return nil, nil
			}
			return nil, err
		}

		for _, item := range resp.Gateways {
			d.StreamListItem(ctx, item)
			count++

			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		if len(resp.Gateways) == 0 || uint64(count) >= uint64(resp.TotalCount) {
			break
		}
		req.Page = scw.Int32Ptr(*req.Page + 1)
	}

	return nil, nil
}
