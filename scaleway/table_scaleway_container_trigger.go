package scaleway

import (
	"context"

	container "github.com/scaleway/scaleway-sdk-go/api/container/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

//// TABLE DEFINITION

func tableScalewayContainerTrigger(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:              "scaleway_container_trigger",
		Description:       "A container trigger invokes a container from a cron schedule or a messaging source. Cron body and headers and credentials are not exposed.",
		GetMatrixItemFunc: BuildRegionList,
		List: &plugin.ListConfig{
			Hydrate:       listContainerTriggers,
			ParentHydrate: listContainerNamespaces,
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
				{
					Name:    "container_id",
					Require: plugin.Optional,
				},
			},
		},
		Columns: []*plugin.Column{
			{
				Name:        "id",
				Description: "The unique identifier of the trigger.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID"),
			},
			{
				Name:        "name",
				Description: "The name of the trigger.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "description",
				Description: "The description of the trigger.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "namespace_id",
				Description: "The ID of the namespace the trigger belongs to.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("NamespaceID"),
			},
			{
				Name:        "container_id",
				Description: "The ID of the container the trigger invokes.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ContainerID"),
			},
			{
				Name:        "source_type",
				Description: "The type of the trigger source.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("SourceType").Transform(transform.ToString),
			},
			{
				Name:        "status",
				Description: "The status of the trigger.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Status").Transform(transform.ToString),
			},
			{
				Name:        "error_message",
				Description: "The error message of the trigger, if any.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "http_path",
				Description: "The HTTP path the trigger calls on the container.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("HTTPPath"),
			},
			{
				Name:        "http_method",
				Description: "The HTTP method the trigger uses.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("HTTPMethod"),
			},
			{
				Name:        "cron_config",
				Description: "The cron schedule and timezone, without the request body and headers.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "sqs_config",
				Description: "The SQS configuration, without the access key ID.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "nats_config",
				Description: "The NATS configuration, with user info removed from server URLs.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "srn",
				Description: "The Scaleway resource name of the trigger.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Srn"),
			},
			{
				Name:        "tags",
				Description: "A list of tags attached to the trigger.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "created_at",
				Description: "The time when the trigger was created.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "updated_at",
				Description: "The time when the trigger was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
			},

			// Scaleway standard columns
			{
				Name:        "region",
				Description: "The region where the trigger resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Region").Transform(transform.ToString),
			},
			{
				Name:        "project_id",
				Description: "The ID of the project where the trigger resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ProjectID"),
			},
			{
				Name:        "organization_id",
				Description: "The ID of the organization where the trigger resides.",
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

type containerTriggerRow struct {
	container.Trigger
	NamespaceID    string
	HTTPPath       string
	HTTPMethod     string
	CronConfig     *containerTriggerCronConfig
	SqsConfig      *containerTriggerSqsConfig
	NatsConfig     *containerTriggerNatsConfig
	ProjectID      string
	OrganizationID string
}

type containerTriggerCronConfig struct {
	Schedule string `json:"schedule"`
	Timezone string `json:"timezone"`
}

type containerTriggerSqsConfig struct {
	Region   string `json:"region"`
	Endpoint string `json:"endpoint"`
	QueueURL string `json:"queue_url"`
}

type containerTriggerNatsConfig struct {
	ServerURLs []string `json:"server_urls"`
	Subject    string   `json:"subject"`
}

// newContainerTriggerRow copies the trigger configs without the cron body and headers, the SQS access key ID
// and URL user info.
func newContainerTriggerRow(t *container.Trigger, ns *container.Namespace) containerTriggerRow {
	row := containerTriggerRow{Trigger: *t, NamespaceID: ns.ID, ProjectID: ns.ProjectID, OrganizationID: ns.OrganizationID}
	if c := t.DestinationConfig; c != nil {
		row.HTTPPath, row.HTTPMethod = c.HTTPPath, c.HTTPMethod.String()
	}
	if c := t.CronConfig; c != nil {
		row.CronConfig = &containerTriggerCronConfig{Schedule: c.Schedule, Timezone: c.Timezone}
	}
	if c := t.SqsConfig; c != nil {
		row.SqsConfig = &containerTriggerSqsConfig{Region: c.Region.String(), Endpoint: redactURLUserInfo(c.Endpoint), QueueURL: redactURLUserInfo(c.QueueURL)}
	}
	if c := t.NatsConfig; c != nil {
		urls := make([]string, 0, len(c.ServerURLs))
		for _, u := range c.ServerURLs {
			urls = append(urls, redactURLUserInfo(u))
		}
		row.NatsConfig = &containerTriggerNatsConfig{ServerURLs: urls, Subject: c.Subject}
	}
	return row
}

//// LIST FUNCTION

func listContainerTriggers(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	matrixLoc := d.EqualsQualString("region")
	parent := h.Item.(*container.Namespace)

	if parent == nil || parent.Region.String() != matrixLoc {
		return nil, nil
	}
	if id := d.EqualsQualString("namespace_id"); id != "" && id != parent.ID {
		return nil, nil
	}

	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_container_trigger.listContainerTriggers", "connection_error", err)
		return nil, err
	}

	api := container.NewAPI(client)

	req := &container.ListTriggersRequest{
		Region:      parent.Region,
		NamespaceID: scw.StringPtr(parent.ID),
		Page:        scw.Int32Ptr(1),
		PageSize:    scw.Uint32Ptr(listPageSize(d)),
	}

	var count int

	for {
		resp, err := api.ListTriggers(req)
		if err != nil {
			plugin.Logger(ctx).Error("scaleway_container_trigger.listContainerTriggers", "query_error", err)
			if is403Error(err) || isNotFoundOrUnsupportedZoneError(err) {
				return nil, nil
			}
			return nil, err
		}

		for _, item := range resp.Triggers {
			if item == nil {
				continue
			}
			if id := d.EqualsQualString("container_id"); id != "" && id != item.ContainerID {
				continue
			}
			d.StreamListItem(ctx, newContainerTriggerRow(item, parent))
			count++

			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		if len(resp.Triggers) == 0 || uint64(count) >= uint64(resp.TotalCount) {
			break
		}
		req.Page = scw.Int32Ptr(*req.Page + 1)
	}

	return nil, nil
}
