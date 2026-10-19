package scaleway

import (
	"context"

	searchdb "github.com/scaleway/scaleway-sdk-go/api/searchdb/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

//// TABLE DEFINITION

func tableScalewayOpenSearchDeployment(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:              "scaleway_opensearch_deployment",
		Description:       "A Managed OpenSearch deployment. Credentials are not exposed.",
		GetMatrixItemFunc: BuildRegionList,
		List: &plugin.ListConfig{
			Hydrate: listOpenSearchDeployments,
			KeyColumns: []*plugin.KeyColumn{
				{
					Name:    "region",
					Require: plugin.Optional,
				},
				{
					Name:    "project_id",
					Require: plugin.Optional,
				},
			},
		},
		Columns: []*plugin.Column{
			{
				Name:        "id",
				Description: "The unique identifier of the deployment.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID"),
			},
			{
				Name:        "name",
				Description: "The name of the deployment.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "status",
				Description: "The status of the deployment.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Status").Transform(transform.ToString),
			},
			{
				Name:        "version",
				Description: "The OpenSearch version of the deployment.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "node_type",
				Description: "The node type of the deployment.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "node_amount",
				Description: "The number of nodes of the deployment.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "node_count",
				Description: "The number of nodes currently in the deployment.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "volume",
				Description: "The volume of the deployment.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "endpoints",
				Description: "The endpoints of the deployment.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "tags",
				Description: "A list of tags attached to the deployment.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "created_at",
				Description: "The time when the deployment was created.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "updated_at",
				Description: "The time when the deployment was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
			},

			// Scaleway standard columns
			{
				Name:        "region",
				Description: "The region where the deployment resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Region").Transform(transform.ToString),
			},
			{
				Name:        "project_id",
				Description: "The ID of the project where the deployment resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ProjectID"),
			},
			{
				Name:        "organization_id",
				Description: "The ID of the organization where the deployment resides.",
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

func listOpenSearchDeployments(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	region, err := scw.ParseRegion(d.EqualsQualString("region"))
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_opensearch_deployment.listOpenSearchDeployments", "region_parsing_error", err)
		return nil, err
	}

	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_opensearch_deployment.listOpenSearchDeployments", "connection_error", err)
		return nil, err
	}

	api := searchdb.NewAPI(client)

	req := &searchdb.ListDeploymentsRequest{
		Region:   region,
		Page:     scw.Int32Ptr(1),
		PageSize: scw.Uint32Ptr(listPageSize(d)),
	}
	if project := d.EqualsQualString("project_id"); project != "" {
		req.ProjectID = scw.StringPtr(project)
	}

	var count int

	for {
		resp, err := api.ListDeployments(req)
		if err != nil {
			plugin.Logger(ctx).Error("scaleway_opensearch_deployment.listOpenSearchDeployments", "query_error", err)
			if is403Error(err) || isNotFoundOrUnsupportedZoneError(err) {
				return nil, nil
			}
			return nil, err
		}

		for _, item := range resp.Deployments {
			if item == nil {
				continue
			}
			d.StreamListItem(ctx, item)
			count++

			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		if len(resp.Deployments) == 0 || uint64(count) >= uint64(resp.TotalCount) {
			break
		}
		req.Page = scw.Int32Ptr(*req.Page + 1)
	}

	return nil, nil
}
