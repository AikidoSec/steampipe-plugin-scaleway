package scaleway

import (
	"context"

	"github.com/scaleway/scaleway-sdk-go/api/serverless_sqldb/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

//// TABLE DEFINITION

func tableScalewayServerlessSQLDBBackup(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:              "scaleway_serverless_sqldb_backup",
		Description:       "A Serverless SQL Database backup. The download URL is not exposed.",
		GetMatrixItemFunc: BuildRegionList,
		List: &plugin.ListConfig{
			Hydrate:       listServerlessSQLDBBackups,
			ParentHydrate: listServerlessSQLDBDatabases,
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
					Name:    "database_id",
					Require: plugin.Optional,
				},
			},
		},
		Columns: []*plugin.Column{
			{
				Name:        "id",
				Description: "The unique identifier of the backup.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID"),
			},
			{
				Name:        "database_id",
				Description: "The ID of the database the backup belongs to.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("DatabaseID"),
			},
			{
				Name:        "status",
				Description: "The status of the backup.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Status").Transform(transform.ToString),
			},
			{
				Name:        "size",
				Description: "The size of the backup, in bytes.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("Size").Transform(sizeToInt),
			},
			{
				Name:        "db_size",
				Description: "The size of the database when the backup was made, in bytes.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("DbSize").Transform(sizeToInt),
			},
			{
				Name:        "srn",
				Description: "The Scaleway resource name of the backup.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Srn"),
			},
			{
				Name:        "expires_at",
				Description: "The time when the backup expires.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "created_at",
				Description: "The time when the backup was created.",
				Type:        proto.ColumnType_TIMESTAMP,
			},

			// Scaleway standard columns
			{
				Name:        "region",
				Description: "The region where the backup resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Region").Transform(transform.ToString),
			},
			{
				Name:        "project_id",
				Description: "The ID of the project where the backup resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ProjectID"),
			},
			{
				Name:        "organization_id",
				Description: "The ID of the organization where the backup resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("OrganizationID"),
			},

			// Steampipe standard columns
			{
				Name:        "title",
				Description: "Title of the resource.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID"),
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

func listServerlessSQLDBBackups(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	matrixRegion := d.EqualsQualString("region")
	parent := h.Item.(*serverless_sqldb.Database)

	if parent == nil || parent.Region.String() != matrixRegion {
		return nil, nil
	}
	if id := d.EqualsQualString("database_id"); id != "" && id != parent.ID {
		return nil, nil
	}

	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_serverless_sqldb_backup.listServerlessSQLDBBackups", "connection_error", err)
		return nil, err
	}

	api := serverless_sqldb.NewAPI(client)

	req := &serverless_sqldb.ListDatabaseBackupsRequest{
		Region:     parent.Region,
		DatabaseID: parent.ID,
		Page:       scw.Int32Ptr(1),
		PageSize:   scw.Uint32Ptr(listPageSize(d)),
	}

	var count int

	for {
		resp, err := api.ListDatabaseBackups(req)
		if err != nil {
			plugin.Logger(ctx).Error("scaleway_serverless_sqldb_backup.listServerlessSQLDBBackups", "query_error", err)
			if is403Error(err) || isNotFoundOrUnsupportedZoneError(err) {
				return nil, nil
			}
			return nil, err
		}

		for _, item := range resp.Backups {
			if item == nil {
				continue
			}
			d.StreamListItem(ctx, item)
			count++

			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		if len(resp.Backups) == 0 || uint64(count) >= resp.TotalCount {
			break
		}
		req.Page = scw.Int32Ptr(*req.Page + 1)
	}

	return nil, nil
}
