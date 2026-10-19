package scaleway

import (
	"context"

	rdb "github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

//// TABLE DEFINITION

func tableScalewayRDBUser(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:              "scaleway_rdb_user",
		Description:       "An RDB user is a database user of a Database Instance. Only the name and admin flag are exposed.",
		GetMatrixItemFunc: BuildRegionList,
		List: &plugin.ListConfig{
			Hydrate:       listRDBUsers,
			ParentHydrate: listRDBInstances,
			KeyColumns: []*plugin.KeyColumn{
				{
					Name:    "region",
					Require: plugin.Optional,
				},
				{
					Name:    "instance_id",
					Require: plugin.Optional,
				},
			},
		},
		Columns: []*plugin.Column{
			{
				Name:        "id",
				Description: "A composite identifier (instance_id/name).",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID"),
			},
			{
				Name:        "name",
				Description: "The name of the user.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "instance_id",
				Description: "The ID of the Database Instance the user belongs to.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("InstanceID"),
			},
			{
				Name:        "is_admin",
				Description: "Indicates whether the user is an administrator.",
				Type:        proto.ColumnType_BOOL,
			},

			// Scaleway standard columns
			{
				Name:        "region",
				Description: "The region where the user resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Region").Transform(transform.ToString),
			},
			{
				Name:        "project_id",
				Description: "The ID of the project where the user resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ProjectID"),
			},
			{
				Name:        "organization_id",
				Description: "The ID of the organization where the user resides.",
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

type rdbUserRow struct {
	rdb.User
	ID             string
	InstanceID     string
	Region         scw.Region
	ProjectID      string
	OrganizationID string
}

//// LIST FUNCTION

func listRDBUsers(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	matrixRegion := d.EqualsQualString("region")
	parent := h.Item.(*rdb.Instance)

	if parent.Region.String() != matrixRegion {
		return nil, nil
	}
	if id := d.EqualsQualString("instance_id"); id != "" && id != parent.ID {
		return nil, nil
	}

	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_rdb_user.listRDBUsers", "connection_error", err)
		return nil, err
	}

	api := rdb.NewAPI(client)

	req := &rdb.ListUsersRequest{
		Region:     parent.Region,
		InstanceID: parent.ID,
		Page:       scw.Int32Ptr(1),
		PageSize:   scw.Uint32Ptr(listPageSize(d)),
	}

	var count int

	for {
		resp, err := api.ListUsers(req)
		if err != nil {
			plugin.Logger(ctx).Error("scaleway_rdb_user.listRDBUsers", "query_error", err)
			if isNotFoundOrUnsupportedZoneError(err) {
				return nil, nil
			}
			return nil, err
		}

		for _, item := range resp.Users {
			d.StreamListItem(ctx, rdbUserRow{*item, parent.ID + "/" + item.Name, parent.ID, parent.Region, parent.ProjectID, parent.OrganizationID})
			count++

			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		if len(resp.Users) == 0 || resp.TotalCount == uint32(count) {
			break
		}
		req.Page = scw.Int32Ptr(*req.Page + 1)
	}

	return nil, nil
}
