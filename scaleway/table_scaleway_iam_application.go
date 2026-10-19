package scaleway

import (
	"context"

	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

//// TABLE DEFINITION

func tableScalewayIamApplication(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "scaleway_iam_application",
		Description: "Applications are non-human identities that own API keys and can be attached to IAM policies.",
		List: &plugin.ListConfig{
			Hydrate: listIamApplications,
		},
		Get: &plugin.GetConfig{
			Hydrate:    getIamApplication,
			KeyColumns: plugin.SingleColumn("id"),
		},
		Columns: []*plugin.Column{
			{
				Name:        "id",
				Description: "The unique identifier of the application.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID"),
			},
			{
				Name:        "name",
				Description: "The name of the application.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "description",
				Description: "The description of the application.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "editable",
				Description: "Whether the application can be edited.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "managed",
				Description: "Whether the application is managed by Scaleway.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "nb_api_keys",
				Description: "The number of API keys owned by the application.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("NbAPIKeys"),
			},
			{
				Name:        "tags",
				Description: "A list of tags attached to the application.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "created_at",
				Description: "The time when the application was created.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "updated_at",
				Description: "The time when the application was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
			},

			// Scaleway standard columns
			{
				Name:        "organization",
				Description: "The ID of the organization where the application resides.",
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

func listIamApplications(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	// Create client
	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_iam_application.listIamApplications", "connection_error", err)
		return nil, err
	}

	iamApi := iam.NewAPI(client)

	req := &iam.ListApplicationsRequest{
		Page: scw.Int32Ptr(1),
	}
	if organisationId := GetConfig(d.Connection).OrganizationID; organisationId != nil {
		req.OrganizationID = *organisationId
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
		resp, err := iamApi.ListApplications(req)
		if err != nil {
			plugin.Logger(ctx).Error("scaleway_iam_application.listIamApplications", "query_error", err)
			return nil, err
		}

		for _, app := range resp.Applications {
			d.StreamListItem(ctx, app)

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

func getIamApplication(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	id := d.EqualsQualString("id")
	if id == "" {
		return nil, nil
	}

	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_iam_application.getIamApplication", "connection_error", err)
		return nil, err
	}

	iamApi := iam.NewAPI(client)

	data, err := iamApi.GetApplication(&iam.GetApplicationRequest{
		ApplicationID: id,
	})
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_iam_application.getIamApplication", "query_error", err)
		if is404Error(err) {
			return nil, nil
		}
		return nil, err
	}

	return data, nil
}
