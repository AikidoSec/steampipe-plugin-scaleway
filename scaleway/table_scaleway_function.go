package scaleway

import (
	"context"

	function "github.com/scaleway/scaleway-sdk-go/api/function/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

//// TABLE DEFINITION

func tableScalewayFunction(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:              "scaleway_function",
		Description:       "A Serverless Function is a piece of code run on demand. Environment variables are not exposed.",
		GetMatrixItemFunc: BuildRegionList,
		List: &plugin.ListConfig{
			Hydrate:       listFunctions,
			ParentHydrate: listFunctionNamespaces,
			KeyColumns: []*plugin.KeyColumn{
				{
					Name:    "region",
					Require: plugin.Optional,
				},
				{
					Name:    "project_id",
					Require: plugin.Optional,
				},
				{
					Name:    "namespace_id",
					Require: plugin.Optional,
				},
			},
		},
		Columns: []*plugin.Column{
			{
				Name:        "id",
				Description: "The unique identifier of the function.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID"),
			},
			{
				Name:        "name",
				Description: "The name of the function.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "namespace_id",
				Description: "The ID of the namespace the function belongs to.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("NamespaceID"),
			},
			{
				Name:        "description",
				Description: "The description of the function.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "status",
				Description: "The status of the function.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Status").Transform(transform.ToString),
			},
			{
				Name:        "error_message",
				Description: "The error message of the function, if any.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "build_message",
				Description: "The build message of the function.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "runtime_message",
				Description: "The runtime message of the function.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "runtime",
				Description: "The runtime of the function.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Runtime").Transform(transform.ToString),
			},
			{
				Name:        "handler",
				Description: "The handler of the function.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "privacy",
				Description: "The privacy of the function: public or private.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Privacy").Transform(transform.ToString),
			},
			{
				Name:        "http_option",
				Description: "Indicates whether HTTP is enabled or redirected to HTTPS.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("HTTPOption").Transform(transform.ToString),
			},
			{
				Name:        "sandbox",
				Description: "The sandbox environment of the function.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Sandbox").Transform(transform.ToString),
			},
			{
				Name:        "min_scale",
				Description: "The minimum number of instances.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "max_scale",
				Description: "The maximum number of instances.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "memory_limit",
				Description: "The memory limit of the function, in MB.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "cpu_limit",
				Description: "The CPU limit of the function, in mVCPU.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("CPULimit"),
			},
			{
				Name:        "timeout",
				Description: "The execution timeout of the function, in seconds.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("Timeout").Transform(durationToSeconds),
			},
			{
				Name:        "domain_name",
				Description: "The default domain name of the function.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "private_network_id",
				Description: "The ID of the private network the function is attached to.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("PrivateNetworkID"),
			},
			{
				Name:        "srn",
				Description: "The Scaleway resource name of the function.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Srn"),
			},
			{
				Name:        "tags",
				Description: "A list of tags attached to the function.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "created_at",
				Description: "The time when the function was created.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "updated_at",
				Description: "The time when the function was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "ready_at",
				Description: "The time when the function became ready.",
				Type:        proto.ColumnType_TIMESTAMP,
			},

			// Scaleway standard columns
			{
				Name:        "region",
				Description: "The region where the function resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Region").Transform(transform.ToString),
			},
			{
				Name:        "project_id",
				Description: "The ID of the project where the function resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ProjectID"),
			},
			{
				Name:        "organization_id",
				Description: "The ID of the organization where the function resides.",
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

type functionRow struct {
	function.Function
	ProjectID      string
	OrganizationID string
}

//// LIST FUNCTION

func listFunctions(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	matrixLoc := d.EqualsQualString("region")
	parent := h.Item.(*function.Namespace)

	if parent == nil || parent.Region.String() != matrixLoc {
		return nil, nil
	}
	if id := d.EqualsQualString("namespace_id"); id != "" && id != parent.ID {
		return nil, nil
	}

	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_function.listFunctions", "connection_error", err)
		return nil, err
	}

	api := function.NewAPI(client)

	req := &function.ListFunctionsRequest{
		Region:      parent.Region,
		NamespaceID: parent.ID,
		Page:        scw.Int32Ptr(1),
		PageSize:    scw.Uint32Ptr(listPageSize(d)),
	}

	var count int

	for {
		resp, err := api.ListFunctions(req)
		if err != nil {
			plugin.Logger(ctx).Error("scaleway_function.listFunctions", "query_error", err)
			if is403Error(err) || isNotFoundOrUnsupportedZoneError(err) {
				return nil, nil
			}
			return nil, err
		}

		for _, item := range resp.Functions {
			if item == nil {
				continue
			}
			d.StreamListItem(ctx, functionRow{*item, parent.ProjectID, parent.OrganizationID})
			count++

			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		if len(resp.Functions) == 0 || uint64(count) >= uint64(resp.TotalCount) {
			break
		}
		req.Page = scw.Int32Ptr(*req.Page + 1)
	}

	return nil, nil
}
