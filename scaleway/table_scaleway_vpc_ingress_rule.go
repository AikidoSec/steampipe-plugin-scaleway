package scaleway

import (
	"context"

	"github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

//// TABLE DEFINITION

func tableScalewayVPCIngressRule(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:              "scaleway_vpc_ingress_rule",
		Description:       "A VPC ingress rule redirects incoming traffic from a source network to a resource in a private network.",
		GetMatrixItemFunc: BuildRegionList,
		List: &plugin.ListConfig{
			Hydrate: listVPCIngressRules,
			KeyColumns: []*plugin.KeyColumn{
				{
					Name:    "region",
					Require: plugin.Optional,
				},
				{
					Name:    "vpc_id",
					Require: plugin.Optional,
				},
			},
		},
		Columns: []*plugin.Column{
			{
				Name:        "id",
				Description: "The unique identifier of the ingress rule.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID"),
			},
			{
				Name:        "vpc_id",
				Description: "The ID of the VPC the rule belongs to.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("VpcID"),
			},
			{
				Name:        "description",
				Description: "The description of the rule.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Description"),
			},
			{
				Name:        "is_ipv6",
				Description: "Indicates whether the rule applies to IPv6 traffic.",
				Type:        proto.ColumnType_BOOL,
				Transform:   transform.FromField("IsIPv6"),
			},
			{
				Name:        "source",
				Description: "The source network the rule applies to.",
				Type:        proto.ColumnType_CIDR,
				Transform:   transform.FromField("Source").Transform(ipNetToString),
			},
			{
				Name:        "nexthop_resource_ip",
				Description: "The IP of the local resource ingress traffic is redirected to.",
				Type:        proto.ColumnType_IPADDR,
				Transform:   transform.FromField("NexthopResourceIP").Transform(safeIPAddrToString),
			},
			{
				Name:        "nexthop_private_network_id",
				Description: "The ID of the private network the destination resource is in.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("NexthopPrivateNetworkID"),
			},
			{
				Name:        "tags",
				Description: "A list of tags attached to the ingress rule.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Tags"),
			},
			{
				Name:        "srn",
				Description: "The Scaleway resource name of the ingress rule.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Srn"),
			},
			{
				Name:        "created_at",
				Description: "The time when the ingress rule was created.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("CreatedAt"),
			},
			{
				Name:        "updated_at",
				Description: "The time when the ingress rule was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("UpdatedAt"),
			},

			// Scaleway standard columns
			{
				Name:        "region",
				Description: "The region where the ingress rule resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Region").Transform(transform.ToString),
			},
			{
				Name:        "project_id",
				Description: "The ID of the project where the ingress rule resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ProjectID"),
			},
			{
				Name:        "organization_id",
				Description: "The ID of the organization where the ingress rule resides.",
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

//// LIST FUNCTION

func listVPCIngressRules(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	region, err := scw.ParseRegion(d.EqualsQualString("region"))
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_vpc_ingress_rule.listVPCIngressRules", "region_parsing_error", err)
		return nil, err
	}

	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_vpc_ingress_rule.listVPCIngressRules", "connection_error", err)
		return nil, err
	}

	api := vpc.NewAPI(client)

	req := &vpc.ListIngressRulesRequest{
		Region:   region,
		Page:     scw.Int32Ptr(1),
		PageSize: scw.Uint32Ptr(listPageSize(d)),
	}
	if v := d.EqualsQualString("vpc_id"); v != "" {
		req.VpcID = scw.StringPtr(v)
	}

	var count int

	for {
		resp, err := api.ListIngressRules(req)
		if err != nil {
			plugin.Logger(ctx).Error("scaleway_vpc_ingress_rule.listVPCIngressRules", "query_error", err)
			if isNotFoundOrUnsupportedZoneError(err) {
				return nil, nil
			}
			return nil, err
		}

		for _, item := range resp.Rules {
			d.StreamListItem(ctx, item)
			count++

			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		if len(resp.Rules) == 0 || uint64(count) >= uint64(resp.TotalCount) {
			break
		}
		req.Page = scw.Int32Ptr(*req.Page + 1)
	}

	return nil, nil
}
