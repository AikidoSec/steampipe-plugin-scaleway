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

func tableScalewayContainer(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:              "scaleway_container",
		Description:       "A Serverless Container runs a container image on demand. Environment variables, command and arguments are not exposed.",
		GetMatrixItemFunc: BuildRegionList,
		List: &plugin.ListConfig{
			Hydrate:       listContainers,
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
			},
		},
		Columns: []*plugin.Column{
			{
				Name:        "id",
				Description: "The unique identifier of the container.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID"),
			},
			{
				Name:        "name",
				Description: "The name of the container.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "namespace_id",
				Description: "The ID of the namespace the container belongs to.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("NamespaceID"),
			},
			{
				Name:        "description",
				Description: "The description of the container.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "status",
				Description: "The status of the container.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Status").Transform(transform.ToString),
			},
			{
				Name:        "error_message",
				Description: "The error message of the container, if any.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "image",
				Description: "The image reference of the container.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "protocol",
				Description: "The protocol the container uses.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Protocol").Transform(transform.ToString),
			},
			{
				Name:        "port",
				Description: "The port the container listens on.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "privacy",
				Description: "The privacy of the container: public or private.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Privacy").Transform(transform.ToString),
			},
			{
				Name:        "https_connections_only",
				Description: "Indicates whether the container only accepts HTTPS connections.",
				Type:        proto.ColumnType_BOOL,
				Transform:   transform.FromField("HTTPSConnectionsOnly"),
			},
			{
				Name:        "sandbox",
				Description: "The sandbox environment of the container.",
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
				Name:        "memory_limit_bytes",
				Description: "The memory limit of the container, in bytes.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("MemoryLimitBytes").Transform(sizeToInt),
			},
			{
				Name:        "mvcpu_limit",
				Description: "The CPU limit of the container, in mVCPU.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("MvcpuLimit"),
			},
			{
				Name:        "local_storage_limit_bytes",
				Description: "The local storage limit of the container, in bytes.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("LocalStorageLimitBytes").Transform(sizeToInt),
			},
			{
				Name:        "timeout",
				Description: "The execution timeout of the container, in seconds.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("Timeout").Transform(durationToSeconds),
			},
			{
				Name:        "scaling_option",
				Description: "The scaling option of the container.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "liveness_probe",
				Description: "The liveness probe of the container.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "startup_probe",
				Description: "The startup probe of the container.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "private_network_id",
				Description: "The ID of the private network the container is attached to.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("PrivateNetworkID"),
			},
			{
				Name:        "public_endpoint",
				Description: "The public endpoint of the container.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "private_endpoint",
				Description: "The private endpoint of the container.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "default_public_endpoint_enabled",
				Description: "Indicates whether the default public endpoint is enabled.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "srn",
				Description: "The Scaleway resource name of the container.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Srn"),
			},
			{
				Name:        "tags",
				Description: "A list of tags attached to the container.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "created_at",
				Description: "The time when the container was created.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "updated_at",
				Description: "The time when the container was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
			},

			// Scaleway standard columns
			{
				Name:        "region",
				Description: "The region where the container resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Region").Transform(transform.ToString),
			},
			{
				Name:        "project_id",
				Description: "The ID of the project where the container resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ProjectID"),
			},
			{
				Name:        "organization_id",
				Description: "The ID of the organization where the container resides.",
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

type containerRow struct {
	container.Container
	ProjectID      string
	OrganizationID string
}

//// LIST FUNCTION

func listContainers(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
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
		plugin.Logger(ctx).Error("scaleway_container.listContainers", "connection_error", err)
		return nil, err
	}

	api := container.NewAPI(client)

	req := &container.ListContainersRequest{
		Region:      parent.Region,
		NamespaceID: scw.StringPtr(parent.ID),
		Page:        scw.Int32Ptr(1),
		PageSize:    scw.Uint32Ptr(listPageSize(d)),
	}

	var count int

	for {
		resp, err := api.ListContainers(req)
		if err != nil {
			plugin.Logger(ctx).Error("scaleway_container.listContainers", "query_error", err)
			if is403Error(err) || isNotFoundOrUnsupportedZoneError(err) {
				return nil, nil
			}
			return nil, err
		}

		for _, item := range resp.Containers {
			if item == nil {
				continue
			}
			d.StreamListItem(ctx, containerRow{*item, parent.ProjectID, parent.OrganizationID})
			count++

			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		if len(resp.Containers) == 0 || uint64(count) >= uint64(resp.TotalCount) {
			break
		}
		req.Page = scw.Int32Ptr(*req.Page + 1)
	}

	return nil, nil
}
