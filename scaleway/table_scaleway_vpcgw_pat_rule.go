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

func tableScalewayVPCGWPATRule(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:              "scaleway_vpcgw_pat_rule",
		Description:       "A PAT rule forwards a public port of a Public Gateway to a private IP address and port.",
		GetMatrixItemFunc: BuildZoneList,
		List: &plugin.ListConfig{
			Hydrate:       listVPCGWPATRules,
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
				Description: "The unique identifier of the PAT rule.",
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
				Name:        "public_port",
				Description: "The public port of the rule.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("PublicPort"),
			},
			{
				Name:        "private_ip",
				Description: "The private IP address traffic is forwarded to.",
				Type:        proto.ColumnType_IPADDR,
				Transform:   transform.FromField("PrivateIP").Transform(safeIPAddrToString),
			},
			{
				Name:        "private_port",
				Description: "The private port traffic is forwarded to.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("PrivatePort"),
			},
			{
				Name:        "protocol",
				Description: "The protocol of the rule.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Protocol").Transform(transform.ToString),
			},
			{
				Name:        "srn",
				Description: "The Scaleway resource name of the PAT rule.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Srn"),
			},
			{
				Name:        "created_at",
				Description: "The time when the PAT rule was created.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("CreatedAt"),
			},
			{
				Name:        "updated_at",
				Description: "The time when the PAT rule was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("UpdatedAt"),
			},

			// Scaleway standard columns
			{
				Name:        "region",
				Description: "The region where the PAT rule resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Zone").Transform(lbRegionFromZone),
			},
			{
				Name:        "zone",
				Description: "The zone where the PAT rule resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Zone").Transform(transform.ToString),
			},
			{
				Name:        "project_id",
				Description: "The ID of the project where the PAT rule resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ProjectID"),
			},
			{
				Name:        "organization_id",
				Description: "The ID of the organization where the PAT rule resides.",
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

type vpcgwPATRuleRow struct {
	vpcgw.PatRule
	ProjectID      string
	OrganizationID string
}

//// LIST FUNCTION

func listVPCGWPATRules(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	matrixLoc := d.EqualsQualString("zone")
	parent := h.Item.(*vpcgw.Gateway)

	if parent.Zone.String() != matrixLoc {
		return nil, nil
	}
	if id := d.EqualsQualString("gateway_id"); id != "" && id != parent.ID {
		return nil, nil
	}

	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_vpcgw_pat_rule.listVPCGWPATRules", "connection_error", err)
		return nil, err
	}

	api := vpcgw.NewAPI(client)

	req := &vpcgw.ListPatRulesRequest{
		Zone:       parent.Zone,
		GatewayIDs: []string{parent.ID},
		Page:       scw.Int32Ptr(1),
		PageSize:   scw.Uint32Ptr(listPageSize(d)),
	}

	var count int

	for {
		resp, err := api.ListPatRules(req)
		if err != nil {
			plugin.Logger(ctx).Error("scaleway_vpcgw_pat_rule.listVPCGWPATRules", "query_error", err)
			if isNotFoundOrUnsupportedZoneError(err) {
				return nil, nil
			}
			return nil, err
		}

		for _, item := range resp.PatRules {
			d.StreamListItem(ctx, vpcgwPATRuleRow{*item, parent.ProjectID, parent.OrganizationID})
			count++

			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		if len(resp.PatRules) == 0 || uint64(count) >= uint64(resp.TotalCount) {
			break
		}
		req.Page = scw.Int32Ptr(*req.Page + 1)
	}

	return nil, nil
}
