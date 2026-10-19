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

func tableScalewayServerlessSQLDBDatabase(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:              "scaleway_serverless_sqldb_database",
		Description:       "A Serverless SQL Database.",
		GetMatrixItemFunc: BuildRegionList,
		List: &plugin.ListConfig{
			Hydrate: listServerlessSQLDBDatabases,
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
				Description: "The unique identifier of the database.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID"),
			},
			{
				Name:        "name",
				Description: "The name of the database.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "status",
				Description: "The status of the database.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Status").Transform(transform.ToString),
			},
			{
				Name:        "endpoint",
				Description: "The endpoint of the database, with user info removed.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Endpoint").Transform(redactURLUserInfoTransform),
			},
			{
				Name:        "started",
				Description: "Indicates whether the database is started.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "cpu_min",
				Description: "The minimum CPU capacity of the database.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("CPUMin"),
			},
			{
				Name:        "cpu_max",
				Description: "The maximum CPU capacity of the database.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("CPUMax"),
			},
			{
				Name:        "cpu_current",
				Description: "The current CPU capacity of the database.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("CPUCurrent"),
			},
			{
				Name:        "engine_major_version",
				Description: "The major version of the PostgreSQL engine.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "version",
				Description: "The engine version of the database.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "srn",
				Description: "The Scaleway resource name of the database.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Srn"),
			},
			{
				Name:        "created_at",
				Description: "The time when the database was created.",
				Type:        proto.ColumnType_TIMESTAMP,
			},

			// Scaleway standard columns
			{
				Name:        "region",
				Description: "The region where the database resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Region").Transform(transform.ToString),
			},
			{
				Name:        "project_id",
				Description: "The ID of the project where the database resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ProjectID"),
			},
			{
				Name:        "organization_id",
				Description: "The ID of the organization where the database resides.",
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

func listServerlessSQLDBDatabases(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	region, err := scw.ParseRegion(d.EqualsQualString("region"))
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_serverless_sqldb_database.listServerlessSQLDBDatabases", "region_parsing_error", err)
		return nil, err
	}

	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_serverless_sqldb_database.listServerlessSQLDBDatabases", "connection_error", err)
		return nil, err
	}

	api := serverless_sqldb.NewAPI(client)

	req := &serverless_sqldb.ListDatabasesRequest{
		Region:   region,
		Page:     scw.Int32Ptr(1),
		PageSize: scw.Uint32Ptr(listPageSize(d)),
	}
	if project := d.EqualsQualString("project_id"); project != "" {
		req.ProjectID = project
	}

	var count int

	for {
		resp, err := api.ListDatabases(req)
		if err != nil {
			plugin.Logger(ctx).Error("scaleway_serverless_sqldb_database.listServerlessSQLDBDatabases", "query_error", err)
			if is403Error(err) || isNotFoundOrUnsupportedZoneError(err) {
				return nil, nil
			}
			return nil, err
		}

		for _, item := range resp.Databases {
			if item == nil {
				continue
			}
			d.StreamListItem(ctx, item)
			count++

			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		if len(resp.Databases) == 0 || uint64(count) >= uint64(resp.TotalCount) {
			break
		}
		req.Page = scw.Int32Ptr(*req.Page + 1)
	}

	return nil, nil
}
