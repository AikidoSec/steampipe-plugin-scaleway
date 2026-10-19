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

func tableScalewayInterlinkLink(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:              "scaleway_interlink_link",
		Description:       "An InterLink link connects a VPC to an on-premises network or a partner. The partner pairing key is never exposed.",
		GetMatrixItemFunc: BuildRegionList,
		List: &plugin.ListConfig{
			Hydrate: listInterlinkLinks,
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
				Description: "The unique identifier of the link.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID"),
			},
			{
				Name:        "name",
				Description: "The name of the link.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Name"),
			},
			{
				Name:        "status",
				Description: "The status of the link.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Status").Transform(transform.ToString),
			},
			{
				Name:        "pop_id",
				Description: "The ID of the point of presence hosting the link.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("PopID"),
			},
			{
				Name:        "bandwidth_mbps",
				Description: "The bandwidth of the link, in Mbps.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("BandwidthMbps"),
			},
			{
				Name:        "bgp_v4_status",
				Description: "The status of the IPv4 BGP session.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("BgpV4Status").Transform(transform.ToString),
			},
			{
				Name:        "bgp_v6_status",
				Description: "The status of the IPv6 BGP session.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("BgpV6Status").Transform(transform.ToString),
			},
			{
				Name:        "vpc_id",
				Description: "The ID of the VPC the link is attached to.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("VpcID"),
			},
			{
				Name:        "routing_policy_v4_id",
				Description: "The ID of the IPv4 routing policy of the link.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("RoutingPolicyV4ID"),
			},
			{
				Name:        "routing_policy_v6_id",
				Description: "The ID of the IPv6 routing policy of the link.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("RoutingPolicyV6ID"),
			},
			{
				Name:        "enable_route_propagation",
				Description: "Indicates whether route propagation is enabled.",
				Type:        proto.ColumnType_BOOL,
				Transform:   transform.FromField("EnableRoutePropagation"),
			},
			{
				Name:        "vlan",
				Description: "The VLAN of the link.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("Vlan"),
			},
			{
				Name:        "scw_bgp_config",
				Description: "The Scaleway side BGP configuration.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("ScwBgpConfig"),
			},
			{
				Name:        "peer_bgp_config",
				Description: "The peer side BGP configuration.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("PeerBgpConfig"),
			},
			{
				Name:        "partner",
				Description: "The partner of the link, without its pairing key.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Partner").Transform(interlinkPartnerWithoutKey),
			},
			{
				Name:        "self",
				Description: "The self-hosted connection of the link.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Self"),
			},
			{
				Name:        "tags",
				Description: "A list of tags attached to the link.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Tags"),
			},
			{
				Name:        "srn",
				Description: "The Scaleway resource name of the link.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Srn"),
			},
			{
				Name:        "created_at",
				Description: "The time when the link was created.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("CreatedAt"),
			},
			{
				Name:        "updated_at",
				Description: "The time when the link was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("UpdatedAt"),
			},

			// Scaleway standard columns
			{
				Name:        "region",
				Description: "The region where the link resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Region").Transform(transform.ToString),
			},
			{
				Name:        "project_id",
				Description: "The ID of the project where the link resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ProjectID"),
			},
			{
				Name:        "organization_id",
				Description: "The ID of the organization where the link resides.",
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

func listInterlinkLinks(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	region, err := scw.ParseRegion(d.EqualsQualString("region"))
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_interlink_link.listInterlinkLinks", "region_parsing_error", err)
		return nil, err
	}

	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_interlink_link.listInterlinkLinks", "connection_error", err)
		return nil, err
	}

	api := interlink.NewAPI(client)

	req := &interlink.ListLinksRequest{
		Region:   region,
		Page:     scw.Int32Ptr(1),
		PageSize: scw.Uint32Ptr(listPageSize(d)),
	}
	if v := d.EqualsQualString("vpc_id"); v != "" {
		req.VpcID = scw.StringPtr(v)
	}

	var count int

	for {
		resp, err := api.ListLinks(req)
		if err != nil {
			plugin.Logger(ctx).Error("scaleway_interlink_link.listInterlinkLinks", "query_error", err)
			if isNotFoundOrUnsupportedZoneError(err) {
				return nil, nil
			}
			return nil, err
		}

		for _, item := range resp.Links {
			d.StreamListItem(ctx, item)
			count++

			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		if len(resp.Links) == 0 || uint64(count) >= uint64(resp.TotalCount) {
			break
		}
		req.Page = scw.Int32Ptr(*req.Page + 1)
	}

	return nil, nil
}

//// TRANSFORM FUNCTIONS

// interlinkPartnerWithoutKey returns the partner host of a link with the pairing key removed.
func interlinkPartnerWithoutKey(_ context.Context, d *transform.TransformData) (interface{}, error) {
	partner, ok := d.Value.(*interlink.PartnerHost)
	if !ok || partner == nil {
		return nil, nil
	}
	redacted := *partner
	redacted.PairingKey = ""
	return &redacted, nil
}
