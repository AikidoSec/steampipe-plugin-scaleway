package scaleway

import (
	"context"

	"github.com/scaleway/scaleway-sdk-go/api/vpcgw/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

//// TABLE DEFINITION

func tableScalewayVPCGWGateway(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:              "scaleway_vpcgw_gateway",
		Description:       "A Public Gateway provides NAT, bastion and routing services for private networks.",
		GetMatrixItemFunc: BuildZoneList,
		List: &plugin.ListConfig{
			Hydrate: listVPCGWGateways,
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
				Description: "The unique identifier of the public gateway.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID"),
			},
			{
				Name:        "name",
				Description: "The name of the public gateway.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Name"),
			},
			{
				Name:        "type",
				Description: "The type of the gateway.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Type"),
			},
			{
				Name:        "status",
				Description: "The status of the gateway.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Status").Transform(transform.ToString),
			},
			{
				Name:        "bandwidth",
				Description: "The bandwidth of the gateway, in bits per second.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("Bandwidth"),
			},
			{
				Name:        "ipv4",
				Description: "The public IPv4 address attached to the gateway.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("IPv4"),
			},
			{
				Name:        "gateway_networks",
				Description: "The private network attachments of the gateway.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("GatewayNetworks"),
			},
			{
				Name:        "version",
				Description: "The version of the gateway.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Version"),
			},
			{
				Name:        "can_upgrade_to",
				Description: "The version the gateway can be upgraded to.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("CanUpgradeTo"),
			},
			{
				Name:        "bastion_enabled",
				Description: "Indicates whether the SSH bastion is enabled.",
				Type:        proto.ColumnType_BOOL,
				Transform:   transform.FromField("BastionEnabled"),
			},
			{
				Name:        "bastion_port",
				Description: "The port of the SSH bastion.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("BastionPort"),
			},
			{
				Name:        "bastion_allowed_ips",
				Description: "The IP ranges allowed to connect to the SSH bastion.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("BastionAllowedIPs"),
			},
			{
				Name:        "smtp_enabled",
				Description: "Indicates whether SMTP traffic is allowed.",
				Type:        proto.ColumnType_BOOL,
				Transform:   transform.FromField("SMTPEnabled"),
			},
			{
				Name:        "is_legacy",
				Description: "Indicates whether the gateway is a legacy gateway.",
				Type:        proto.ColumnType_BOOL,
				Transform:   transform.FromField("IsLegacy"),
			},
			{
				Name:        "tags",
				Description: "A list of tags attached to the public gateway.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Tags"),
			},
			{
				Name:        "srn",
				Description: "The Scaleway resource name of the public gateway.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Srn"),
			},
			{
				Name:        "created_at",
				Description: "The time when the public gateway was created.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("CreatedAt"),
			},
			{
				Name:        "updated_at",
				Description: "The time when the public gateway was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("UpdatedAt"),
			},

			// Scaleway standard columns
			{
				Name:        "region",
				Description: "The region where the public gateway resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Zone").Transform(lbRegionFromZone),
			},
			{
				Name:        "zone",
				Description: "The zone where the public gateway resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Zone").Transform(transform.ToString),
			},
			{
				Name:        "project_id",
				Description: "The ID of the project where the public gateway resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ProjectID"),
			},
			{
				Name:        "organization_id",
				Description: "The ID of the organization where the public gateway resides.",
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

func listVPCGWGateways(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	zone, err := scw.ParseZone(d.EqualsQualString("zone"))
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_vpcgw_gateway.listVPCGWGateways", "zone_parsing_error", err)
		return nil, err
	}

	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_vpcgw_gateway.listVPCGWGateways", "connection_error", err)
		return nil, err
	}

	api := vpcgw.NewAPI(client)

	req := &vpcgw.ListGatewaysRequest{
		Zone:     zone,
		Page:     scw.Int32Ptr(1),
		PageSize: scw.Uint32Ptr(listPageSize(d)),
	}

	var count int

	for {
		resp, err := api.ListGateways(req)
		if err != nil {
			plugin.Logger(ctx).Error("scaleway_vpcgw_gateway.listVPCGWGateways", "query_error", err)
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
