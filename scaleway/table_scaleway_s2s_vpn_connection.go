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

func tableScalewayS2SVPNConnection(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:              "scaleway_s2s_vpn_connection",
		Description:       "A site-to-site VPN connection is an IPsec tunnel between a VPN gateway and a customer gateway. The pre-shared key is never exposed.",
		GetMatrixItemFunc: BuildRegionList,
		List: &plugin.ListConfig{
			Hydrate: listS2SVPNConnections,
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
				Description: "The unique identifier of the VPN connection.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID"),
			},
			{
				Name:        "name",
				Description: "The name of the VPN connection.",
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
				Name:        "tunnel_status",
				Description: "The status of the IPsec tunnel.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("TunnelStatus").Transform(transform.ToString),
			},
			{
				Name:        "is_ipv6",
				Description: "Indicates whether the tunnel uses IPv6.",
				Type:        proto.ColumnType_BOOL,
				Transform:   transform.FromField("IsIPv6"),
			},
			{
				Name:        "initiation_policy",
				Description: "The party that initiates the tunnel.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("InitiationPolicy").Transform(transform.ToString),
			},
			{
				Name:        "secret_id",
				Description: "The ID of the Secret Manager secret holding the pre-shared key.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("SecretID"),
			},
			{
				Name:        "secret_revision",
				Description: "The revision of the Secret Manager secret holding the pre-shared key.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("SecretRevision"),
			},
			{
				Name:        "ikev2_ciphers",
				Description: "The IKEv2 ciphers of the connection.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Ikev2Ciphers"),
			},
			{
				Name:        "esp_ciphers",
				Description: "The ESP ciphers of the connection.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("EspCiphers"),
			},
			{
				Name:        "route_propagation_enabled",
				Description: "Indicates whether route propagation is enabled.",
				Type:        proto.ColumnType_BOOL,
				Transform:   transform.FromField("RoutePropagationEnabled"),
			},
			{
				Name:        "vpn_gateway_id",
				Description: "The ID of the VPN gateway of the connection.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("VpnGatewayID"),
			},
			{
				Name:        "customer_gateway_id",
				Description: "The ID of the customer gateway of the connection.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("CustomerGatewayID"),
			},
			{
				Name:        "bgp_status_ipv4",
				Description: "The status of the IPv4 BGP session.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("BgpStatusIPv4").Transform(transform.ToString),
			},
			{
				Name:        "bgp_status_ipv6",
				Description: "The status of the IPv6 BGP session.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("BgpStatusIPv6").Transform(transform.ToString),
			},
			{
				Name:        "bgp_session_ipv4",
				Description: "The IPv4 BGP session configuration.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("BgpSessionIPv4"),
			},
			{
				Name:        "bgp_session_ipv6",
				Description: "The IPv6 BGP session configuration.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("BgpSessionIPv6"),
			},
			{
				Name:        "tags",
				Description: "A list of tags attached to the VPN connection.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Tags"),
			},
			{
				Name:        "srn",
				Description: "The Scaleway resource name of the VPN connection.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Srn"),
			},
			{
				Name:        "created_at",
				Description: "The time when the VPN connection was created.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("CreatedAt"),
			},
			{
				Name:        "updated_at",
				Description: "The time when the VPN connection was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("UpdatedAt"),
			},

			// Scaleway standard columns
			{
				Name:        "region",
				Description: "The region where the VPN connection resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Region").Transform(transform.ToString),
			},
			{
				Name:        "project_id",
				Description: "The ID of the project where the VPN connection resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ProjectID"),
			},
			{
				Name:        "organization_id",
				Description: "The ID of the organization where the VPN connection resides.",
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

func listS2SVPNConnections(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	region, err := scw.ParseRegion(d.EqualsQualString("region"))
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_s2s_vpn_connection.listS2SVPNConnections", "region_parsing_error", err)
		return nil, err
	}

	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_s2s_vpn_connection.listS2SVPNConnections", "connection_error", err)
		return nil, err
	}

	api := s2s_vpn.NewAPI(client)

	req := &s2s_vpn.ListConnectionsRequest{
		Region:   region,
		Page:     scw.Int32Ptr(1),
		PageSize: scw.Uint32Ptr(listPageSize(d)),
	}

	var count int

	for {
		resp, err := api.ListConnections(req)
		if err != nil {
			plugin.Logger(ctx).Error("scaleway_s2s_vpn_connection.listS2SVPNConnections", "query_error", err)
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
