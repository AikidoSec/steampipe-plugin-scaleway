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

func tableScalewayRDBPrivilege(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:              "scaleway_rdb_privilege",
		Description:       "An RDB privilege is the permission a user holds on a database of a Database Instance.",
		GetMatrixItemFunc: BuildRegionList,
		List: &plugin.ListConfig{
			Hydrate:       listRDBPrivileges,
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
				Description: "A composite identifier (instance_id/database_name/user_name).",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID"),
			},
			{
				Name:        "instance_id",
				Description: "The ID of the Database Instance the privilege belongs to.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("InstanceID"),
			},
			{
				Name:        "user_name",
				Description: "The name of the user holding the privilege.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "database_name",
				Description: "The name of the database the privilege applies to.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "permission",
				Description: "The permission granted.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Permission").Transform(transform.ToString),
			},

			// Scaleway standard columns
			{
				Name:        "region",
				Description: "The region where the privilege resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Region").Transform(transform.ToString),
			},
			{
				Name:        "project_id",
				Description: "The ID of the project where the privilege resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ProjectID"),
			},
			{
				Name:        "organization_id",
				Description: "The ID of the organization where the privilege resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("OrganizationID"),
			},

			// Steampipe standard columns
			{
				Name:        "title",
				Description: "Title of the resource.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("UserName"),
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

type rdbPrivilegeRow struct {
	rdb.Privilege
	ID             string
	InstanceID     string
	Region         scw.Region
	ProjectID      string
	OrganizationID string
}

//// LIST FUNCTION

func listRDBPrivileges(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	matrixRegion := d.EqualsQualString("region")
	parent := h.Item.(*rdb.Instance)

	if parent.Region.String() != matrixRegion {
		return nil, nil
	}

	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_rdb_privilege.listRDBPrivileges", "connection_error", err)
		return nil, err
	}

	api := rdb.NewAPI(client)

	req := &rdb.ListPrivilegesRequest{
		Region:     parent.Region,
		InstanceID: parent.ID,
		Page:       scw.Int32Ptr(1),
		PageSize:   scw.Uint32Ptr(listPageSize(d)),
	}

	var count int

	for {
		resp, err := api.ListPrivileges(req)
		if err != nil {
			plugin.Logger(ctx).Error("scaleway_rdb_privilege.listRDBPrivileges", "query_error", err)
			if isNotFoundOrUnsupportedZoneError(err) {
				return nil, nil
			}
			return nil, err
		}

		for _, item := range resp.Privileges {
			d.StreamListItem(ctx, rdbPrivilegeRow{*item, parent.ID + "/" + item.DatabaseName + "/" + item.UserName, parent.ID, parent.Region, parent.ProjectID, parent.OrganizationID})
			count++

			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		if len(resp.Privileges) == 0 || resp.TotalCount == uint32(count) {
			break
		}
		req.Page = scw.Int32Ptr(*req.Page + 1)
	}

	return nil, nil
}
