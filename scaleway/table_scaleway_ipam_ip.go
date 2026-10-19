package scaleway

import (
	"context"

	"github.com/scaleway/scaleway-sdk-go/api/ipam/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

//// TABLE DEFINITION

func tableScalewayIPAMIP(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:              "scaleway_ipam_ip",
		Description:       "An IPAM IP is an IP address managed by the IP Address Management service.",
		GetMatrixItemFunc: BuildRegionList,
		List: &plugin.ListConfig{
			Hydrate: listIPAMIPs,
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
				Description: "The unique identifier of the IPAM IP.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID"),
			},
			{
				Name:        "address",
				Description: "The IP address, as a CIDR.",
				Type:        proto.ColumnType_CIDR,
				Transform:   transform.FromField("Address").Transform(ipNetToString),
			},
			{
				Name:        "is_ipv6",
				Description: "Indicates whether the address is an IPv6 address.",
				Type:        proto.ColumnType_BOOL,
				Transform:   transform.FromField("IsIPv6"),
			},
			{
				Name:        "source",
				Description: "The source the address was allocated from.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Source"),
			},
			{
				Name:        "resource",
				Description: "The resource the address is attached to, if any.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Resource"),
			},
			{
				Name:        "reverses",
				Description: "The reverse DNS entries of the address.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Reverses"),
			},
			{
				Name:        "zone",
				Description: "The zone of the address, for zonal addresses.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Zone").Transform(ipamZoneToString),
			},
			{
				Name:        "tags",
				Description: "A list of tags attached to the IPAM IP.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Tags"),
			},
			{
				Name:        "srn",
				Description: "The Scaleway resource name of the IPAM IP.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Srn"),
			},
			{
				Name:        "created_at",
				Description: "The time when the IPAM IP was created.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("CreatedAt"),
			},
			{
				Name:        "updated_at",
				Description: "The time when the IPAM IP was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("UpdatedAt"),
			},

			// Scaleway standard columns
			{
				Name:        "region",
				Description: "The region where the IPAM IP resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Region").Transform(transform.ToString),
			},
			{
				Name:        "project_id",
				Description: "The ID of the project where the IPAM IP resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ProjectID"),
			},

			// Steampipe standard columns
			{
				Name:        "title",
				Description: "Title of the resource.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Address").Transform(ipNetToString),
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

func listIPAMIPs(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	region, err := scw.ParseRegion(d.EqualsQualString("region"))
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_ipam_ip.listIPAMIPs", "region_parsing_error", err)
		return nil, err
	}

	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_ipam_ip.listIPAMIPs", "connection_error", err)
		return nil, err
	}

	api := ipam.NewAPI(client)

	req := &ipam.ListIPsRequest{
		Region:   region,
		Page:     scw.Int32Ptr(1),
		PageSize: scw.Uint32Ptr(listPageSize(d)),
	}

	var count int

	for {
		resp, err := api.ListIPs(req)
		if err != nil {
			plugin.Logger(ctx).Error("scaleway_ipam_ip.listIPAMIPs", "query_error", err)
			if isNotFoundOrUnsupportedZoneError(err) {
				return nil, nil
			}
			return nil, err
		}

		for _, item := range resp.IPs {
			d.StreamListItem(ctx, item)
			count++

			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		if len(resp.IPs) == 0 || uint64(count) >= uint64(resp.TotalCount) {
			break
		}
		req.Page = scw.Int32Ptr(*req.Page + 1)
	}

	return nil, nil
}

//// TRANSFORM FUNCTIONS

// ipamZoneToString renders an optional zone as text, or nil for regional addresses.
func ipamZoneToString(_ context.Context, d *transform.TransformData) (interface{}, error) {
	zone, ok := d.Value.(*scw.Zone)
	if !ok || zone == nil {
		return nil, nil
	}
	return zone.String(), nil
}
