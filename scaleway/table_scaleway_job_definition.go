package scaleway

import (
	"context"

	jobs "github.com/scaleway/scaleway-sdk-go/api/jobs/v1alpha2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

//// TABLE DEFINITION

func tableScalewayJobDefinition(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:              "scaleway_job_definition",
		Description:       "A Serverless Job definition describes a containerized job. Environment variables, command and arguments are not exposed.",
		GetMatrixItemFunc: BuildRegionList,
		List: &plugin.ListConfig{
			Hydrate: listJobDefinitions,
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
				Description: "The unique identifier of the job definition.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID"),
			},
			{
				Name:        "name",
				Description: "The name of the job definition.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "description",
				Description: "The description of the job definition.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "image_uri",
				Description: "The URI of the container image the job runs.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ImageURI"),
			},
			{
				Name:        "cpu_limit",
				Description: "The CPU limit of the job, in mVCPU.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("CPULimit"),
			},
			{
				Name:        "memory_limit",
				Description: "The memory limit of the job, in MB.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("MemoryLimit"),
			},
			{
				Name:        "local_storage_capacity",
				Description: "The local storage capacity of the job, in MB.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("LocalStorageCapacity"),
			},
			{
				Name:        "job_timeout",
				Description: "The timeout of the job, in seconds.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("JobTimeout").Transform(durationToSeconds),
			},
			{
				Name:        "cron_schedule",
				Description: "The cron schedule of the job.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "retry_policy",
				Description: "The retry policy of the job.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "created_at",
				Description: "The time when the job definition was created.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "updated_at",
				Description: "The time when the job definition was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
			},

			// Scaleway standard columns
			{
				Name:        "region",
				Description: "The region where the job definition resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Region").Transform(transform.ToString),
			},
			{
				Name:        "project_id",
				Description: "The ID of the project where the job definition resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ProjectID"),
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

func listJobDefinitions(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	region, err := scw.ParseRegion(d.EqualsQualString("region"))
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_job_definition.listJobDefinitions", "region_parsing_error", err)
		return nil, err
	}

	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_job_definition.listJobDefinitions", "connection_error", err)
		return nil, err
	}

	api := jobs.NewAPI(client)

	req := &jobs.ListJobDefinitionsRequest{
		Region:   region,
		Page:     scw.Int32Ptr(1),
		PageSize: scw.Uint32Ptr(listPageSize(d)),
	}
	if project := d.EqualsQualString("project_id"); project != "" {
		req.ProjectID = scw.StringPtr(project)
	}

	var count int

	for {
		resp, err := api.ListJobDefinitions(req)
		if err != nil {
			plugin.Logger(ctx).Error("scaleway_job_definition.listJobDefinitions", "query_error", err)
			if is403Error(err) || isNotFoundOrUnsupportedZoneError(err) {
				return nil, nil
			}
			return nil, err
		}

		for _, item := range resp.JobDefinitions {
			if item == nil {
				continue
			}
			d.StreamListItem(ctx, item)
			count++

			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		if len(resp.JobDefinitions) == 0 || uint64(count) >= uint64(resp.TotalCount) {
			break
		}
		req.Page = scw.Int32Ptr(*req.Page + 1)
	}

	return nil, nil
}
