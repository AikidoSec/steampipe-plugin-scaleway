package scaleway

import (
	"context"

	lb "github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

//// TABLE DEFINITION

func tableScalewayLB(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:              "scaleway_lb",
		Description:       "A load balancer distributes incoming traffic across multiple backend servers.",
		GetMatrixItemFunc: BuildZoneList,
		List: &plugin.ListConfig{
			Hydrate: listLBs,
			KeyColumns: []*plugin.KeyColumn{
				{
					Name:    "zone",
					Require: plugin.Optional,
				},
			},
		},
		Get: &plugin.GetConfig{
			Hydrate:    getLB,
			KeyColumns: plugin.AllColumns([]string{"id", "zone"}),
		},
		Columns: []*plugin.Column{
			{
				Name:        "id",
				Description: "The unique identifier of the load balancer.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID"),
			},
			{
				Name:        "name",
				Description: "The name of the load balancer.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "status",
				Description: "The status of the load balancer.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Status").Transform(transform.ToString),
			},
			{
				Name:        "type",
				Description: "The commercial type of the load balancer.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "ip",
				Description: "The IP addresses attached to the load balancer.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("IP"),
			},
			{
				Name:        "tags",
				Description: "A list of tags attached to the load balancer.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "frontend_count",
				Description: "The number of frontends of the load balancer.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "backend_count",
				Description: "The number of backends of the load balancer.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "private_network_count",
				Description: "The number of private networks attached to the load balancer.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "ssl_compatibility_level",
				Description: "The SSL compatibility level of the load balancer.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("SslCompatibilityLevel").Transform(transform.ToString),
			},
			{
				Name:        "created_at",
				Description: "The time when the load balancer was created.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "updated_at",
				Description: "The time when the load balancer was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
			},

			// Scaleway standard columns
			{
				Name:        "region",
				Description: "The region where the load balancer resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Zone").Transform(lbRegionFromZone),
			},
			{
				Name:        "zone",
				Description: "The zone where the load balancer resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Zone").Transform(transform.ToString),
			},
			{
				Name:        "project_id",
				Description: "The ID of the project where the load balancer resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ProjectID"),
			},
			{
				Name:        "organization_id",
				Description: "The ID of the organization where the load balancer resides.",
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

//// TRANSFORM FUNCTIONS

// lbRegionFromZone derives the region name from a scw.Zone, or nil when unparseable.
func lbRegionFromZone(_ context.Context, d *transform.TransformData) (interface{}, error) {
	zone, ok := d.Value.(scw.Zone)
	if !ok {
		return nil, nil
	}
	region, err := zone.Region()
	if err != nil {
		return nil, nil
	}
	return region.String(), nil
}

//// LIST FUNCTION

func listLBs(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	zone := d.EqualsQualString("zone")

	parseZoneData, err := scw.ParseZone(zone)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_lb.listLBs", "zone_parsing_error", err)
		return nil, err
	}

	quals := d.EqualsQuals
	if quals["zone"] != nil && quals["zone"].GetStringValue() != zone {
		return nil, nil
	}

	// Create client
	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_lb.listLBs", "connection_error", err)
		return nil, err
	}

	// Create SDK objects for Scaleway Load Balancer product
	lbApi := lb.NewZonedAPI(client)

	req := &lb.ZonedAPIListLBsRequest{
		Zone: parseZoneData,
		Page: scw.Int32Ptr(1),
	}

	maxResult := int64(100)

	// Reduce the basic request limit down if the user has only requested a small number of rows
	limit := d.QueryContext.Limit
	if d.QueryContext.Limit != nil {
		if *limit < maxResult {
			maxResult = *limit
		}
	}
	req.PageSize = scw.Uint32Ptr(uint32(maxResult))

	var count int

	for {
		resp, err := lbApi.ListLBs(req)
		if err != nil {
			plugin.Logger(ctx).Error("scaleway_lb.listLBs", "query_error", err)
			if isNotFoundOrUnsupportedZoneError(err) {
				return nil, nil
			}
			return nil, err
		}

		for _, item := range resp.LBs {
			d.StreamListItem(ctx, item)

			// Increase the resource count by 1
			count++

			// Context can be cancelled due to manual cancellation or the limit has been hit
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		if resp.TotalCount == uint32(count) {
			break
		}
		req.Page = scw.Int32Ptr(*req.Page + 1)
	}

	return nil, nil
}

//// HYDRATE FUNCTIONS

func getLB(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	zone := d.EqualsQualString("zone")

	parseZoneData, err := scw.ParseZone(zone)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_lb.getLB", "zone_parsing_error", err)
		return nil, err
	}

	id := d.EqualsQualString("id")
	if id == "" {
		return nil, nil
	}

	// Create client
	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_lb.getLB", "connection_error", err)
		return nil, err
	}

	lbApi := lb.NewZonedAPI(client)

	data, err := lbApi.GetLB(&lb.ZonedAPIGetLBRequest{
		Zone: parseZoneData,
		LBID: id,
	})
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_lb.getLB", "query_error", err)
		if is404Error(err) {
			return nil, nil
		}
		return nil, err
	}

	return data, nil
}
