package scaleway

import (
	"context"

	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

//// TABLE DEFINITION

func tableScalewayInstancePlacementGroup(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:              "scaleway_instance_placement_group",
		Description:       "An Instance placement group controls how servers are spread across the hardware of a zone.",
		GetMatrixItemFunc: BuildZoneList,
		List: &plugin.ListConfig{
			Hydrate: listInstancePlacementGroups,
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
				Description: "The unique identifier of the placement group.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID"),
			},
			{
				Name:        "name",
				Description: "The name of the placement group.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Name"),
			},
			{
				Name:        "policy_mode",
				Description: "The policy mode of the placement group.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("PolicyMode").Transform(transform.ToString),
			},
			{
				Name:        "policy_type",
				Description: "The policy type of the placement group.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("PolicyType").Transform(transform.ToString),
			},
			{
				Name:        "policy_respected",
				Description: "Indicates whether the policy is currently respected.",
				Type:        proto.ColumnType_BOOL,
				Transform:   transform.FromField("PolicyRespected"),
			},
			{
				Name:        "tags",
				Description: "A list of tags attached to the placement group.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Tags"),
			},
			{
				Name:        "srn",
				Description: "The Scaleway resource name of the placement group.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Srn"),
			},

			// Scaleway standard columns
			{
				Name:        "region",
				Description: "The region where the placement group resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Zone").Transform(lbRegionFromZone),
			},
			{
				Name:        "zone",
				Description: "The zone where the placement group resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Zone").Transform(transform.ToString),
			},
			{
				Name:        "project_id",
				Description: "The ID of the project where the placement group resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Project"),
			},
			{
				Name:        "organization_id",
				Description: "The ID of the organization where the placement group resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Organization"),
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

func listInstancePlacementGroups(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	zone, err := scw.ParseZone(d.EqualsQualString("zone"))
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_instance_placement_group.listInstancePlacementGroups", "zone_parsing_error", err)
		return nil, err
	}

	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_instance_placement_group.listInstancePlacementGroups", "connection_error", err)
		return nil, err
	}

	api := instance.NewAPI(client)

	req := &instance.ListPlacementGroupsRequest{
		Zone:    zone,
		Page:    scw.Int32Ptr(1),
		PerPage: scw.Uint32Ptr(listPageSize(d)),
	}

	var count int

	for {
		resp, err := api.ListPlacementGroups(req)
		if err != nil {
			plugin.Logger(ctx).Error("scaleway_instance_placement_group.listInstancePlacementGroups", "query_error", err)
			if isNotFoundOrUnsupportedZoneError(err) {
				return nil, nil
			}
			return nil, err
		}

		for _, item := range resp.PlacementGroups {
			d.StreamListItem(ctx, item)
			count++

			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		if len(resp.PlacementGroups) == 0 || uint64(count) >= uint64(resp.TotalCount) {
			break
		}
		req.Page = scw.Int32Ptr(*req.Page + 1)
	}

	return nil, nil
}
