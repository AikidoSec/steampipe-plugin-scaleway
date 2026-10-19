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

func tableScalewayS2SVPNRoutingPolicy(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:              "scaleway_s2s_vpn_routing_policy",
		Description:       "A site-to-site VPN routing policy filters the prefixes exchanged over BGP.",
		GetMatrixItemFunc: BuildRegionList,
		List: &plugin.ListConfig{
			Hydrate: listS2SVPNRoutingPolicies,
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
				Description: "The unique identifier of the routing policy.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID"),
			},
			{
				Name:        "name",
				Description: "The name of the routing policy.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Name"),
			},
			{
				Name:        "is_ipv6",
				Description: "Indicates whether the policy applies to IPv6 prefixes.",
				Type:        proto.ColumnType_BOOL,
				Transform:   transform.FromField("IsIPv6"),
			},
			{
				Name:        "prefix_filter_in",
				Description: "The prefixes accepted from the peer.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("PrefixFilterIn"),
			},
			{
				Name:        "prefix_filter_out",
				Description: "The prefixes announced to the peer.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("PrefixFilterOut"),
			},
			{
				Name:        "tags",
				Description: "A list of tags attached to the routing policy.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Tags"),
			},
			{
				Name:        "srn",
				Description: "The Scaleway resource name of the routing policy.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Srn"),
			},
			{
				Name:        "created_at",
				Description: "The time when the routing policy was created.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("CreatedAt"),
			},
			{
				Name:        "updated_at",
				Description: "The time when the routing policy was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("UpdatedAt"),
			},

			// Scaleway standard columns
			{
				Name:        "region",
				Description: "The region where the routing policy resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Region").Transform(transform.ToString),
			},
			{
				Name:        "project_id",
				Description: "The ID of the project where the routing policy resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ProjectID"),
			},
			{
				Name:        "organization_id",
				Description: "The ID of the organization where the routing policy resides.",
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

func listS2SVPNRoutingPolicies(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	region, err := scw.ParseRegion(d.EqualsQualString("region"))
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_s2s_vpn_routing_policy.listS2SVPNRoutingPolicies", "region_parsing_error", err)
		return nil, err
	}

	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_s2s_vpn_routing_policy.listS2SVPNRoutingPolicies", "connection_error", err)
		return nil, err
	}

	api := s2s_vpn.NewAPI(client)

	req := &s2s_vpn.ListRoutingPoliciesRequest{
		Region:   region,
		Page:     scw.Int32Ptr(1),
		PageSize: scw.Uint32Ptr(listPageSize(d)),
	}

	var count int

	for {
		resp, err := api.ListRoutingPolicies(req)
		if err != nil {
			plugin.Logger(ctx).Error("scaleway_s2s_vpn_routing_policy.listS2SVPNRoutingPolicies", "query_error", err)
			if isNotFoundOrUnsupportedZoneError(err) {
				return nil, nil
			}
			return nil, err
		}

		for _, item := range resp.RoutingPolicies {
			d.StreamListItem(ctx, item)
			count++

			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		if len(resp.RoutingPolicies) == 0 || uint64(count) >= uint64(resp.TotalCount) {
			break
		}
		req.Page = scw.Int32Ptr(*req.Page + 1)
	}

	return nil, nil
}
