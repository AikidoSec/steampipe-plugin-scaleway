package scaleway

import (
	"context"

	edge_services "github.com/scaleway/scaleway-sdk-go/api/edge_services/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

//// TABLE DEFINITION

func tableScalewayEdgeServicesPipeline(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "scaleway_edge_services_pipeline",
		Description: "An Edge Services pipeline chains the stages that process the traffic of a website or application.",
		List: &plugin.ListConfig{
			Hydrate:    listEdgeServicesPipelines,
			KeyColumns: []*plugin.KeyColumn{},
		},
		Columns: []*plugin.Column{
			{
				Name:        "id",
				Description: "The unique identifier of the pipeline.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID"),
			},
			{
				Name:        "name",
				Description: "The name of the pipeline.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Name"),
			},
			{
				Name:        "description",
				Description: "The description of the pipeline.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Description"),
			},
			{
				Name:        "status",
				Description: "The status of the pipeline.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Status").Transform(transform.ToString),
			},
			{
				Name:        "errors",
				Description: "The errors reported on the pipeline.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Errors"),
			},
			{
				Name:        "vpc_endpoint_ids",
				Description: "The IDs of the VPC endpoints used by the pipeline.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("VpcEndpointIDs"),
			},
			{
				Name:        "created_at",
				Description: "The time when the pipeline was created.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("CreatedAt"),
			},
			{
				Name:        "updated_at",
				Description: "The time when the pipeline was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("UpdatedAt"),
			},

			// Scaleway standard columns
			{
				Name:        "project_id",
				Description: "The ID of the project where the pipeline resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ProjectID"),
			},
			{
				Name:        "organization_id",
				Description: "The ID of the organization where the pipeline resides.",
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

func listEdgeServicesPipelines(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {

	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_edge_services_pipeline.listEdgeServicesPipelines", "connection_error", err)
		return nil, err
	}

	api := edge_services.NewAPI(client)

	req := &edge_services.ListPipelinesRequest{
		Page:     scw.Int32Ptr(1),
		PageSize: scw.Uint32Ptr(listPageSize(d)),
	}

	var count int

	for {
		resp, err := api.ListPipelines(req)
		if err != nil {
			plugin.Logger(ctx).Error("scaleway_edge_services_pipeline.listEdgeServicesPipelines", "query_error", err)
			if isNotFoundOrUnsupportedZoneError(err) {
				return nil, nil
			}
			return nil, err
		}

		for _, item := range resp.Pipelines {
			d.StreamListItem(ctx, item)
			count++

			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		if len(resp.Pipelines) == 0 || uint64(count) >= uint64(resp.TotalCount) {
			break
		}
		req.Page = scw.Int32Ptr(*req.Page + 1)
	}

	return nil, nil
}
