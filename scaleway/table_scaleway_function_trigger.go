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

func tableScalewayFunctionTrigger(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:              "scaleway_function_trigger",
		Description:       "A function trigger invokes a function from a messaging source. Credential references and keys are not exposed.",
		GetMatrixItemFunc: BuildRegionList,
		List: &plugin.ListConfig{
			Hydrate:       listFunctionTriggers,
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
				{
					Name:    "function_id",
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
				Name:        "function_id",
				Description: "The ID of the function the trigger invokes.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("FunctionID"),
			},
			{
				Name:        "input_type",
				Description: "The type of the trigger input.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("InputType").Transform(transform.ToString),
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
				Name:        "scw_sqs_config",
				Description: "The Scaleway Messaging SQS configuration, without the credential reference.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "scw_nats_config",
				Description: "The Scaleway Messaging NATS configuration, without the credential reference.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "sqs_config",
				Description: "The external SQS configuration, without the access key and secret key.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "srn",
				Description: "The Scaleway resource name of the trigger.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Srn"),
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

type functionTriggerRow struct {
	function.Trigger
	NamespaceID    string
	ScwSqsConfig   *functionTriggerScwSqsConfig
	ScwNatsConfig  *functionTriggerScwNatsConfig
	SqsConfig      *functionTriggerSqsConfig
	ProjectID      string
	OrganizationID string
}

type functionTriggerScwSqsConfig struct {
	Queue        string `json:"queue"`
	MnqProjectID string `json:"mnq_project_id"`
	MnqRegion    string `json:"mnq_region"`
}

type functionTriggerScwNatsConfig struct {
	Subject          string `json:"subject"`
	MnqNatsAccountID string `json:"mnq_nats_account_id"`
	MnqProjectID     string `json:"mnq_project_id"`
	MnqRegion        string `json:"mnq_region"`
}

type functionTriggerSqsConfig struct {
	Endpoint string `json:"endpoint"`
	QueueURL string `json:"queue_url"`
}

// newFunctionTriggerRow copies the trigger configs without credential references or keys.
func newFunctionTriggerRow(t *function.Trigger, ns *function.Namespace) functionTriggerRow {
	row := functionTriggerRow{Trigger: *t, NamespaceID: ns.ID, ProjectID: ns.ProjectID, OrganizationID: ns.OrganizationID}
	if c := t.ScwSqsConfig; c != nil {
		row.ScwSqsConfig = &functionTriggerScwSqsConfig{Queue: c.Queue, MnqProjectID: c.MnqProjectID, MnqRegion: c.MnqRegion}
	}
	if c := t.ScwNatsConfig; c != nil {
		row.ScwNatsConfig = &functionTriggerScwNatsConfig{Subject: c.Subject, MnqNatsAccountID: c.MnqNatsAccountID, MnqProjectID: c.MnqProjectID, MnqRegion: c.MnqRegion}
	}
	if c := t.SqsConfig; c != nil {
		row.SqsConfig = &functionTriggerSqsConfig{Endpoint: redactURLUserInfo(c.Endpoint), QueueURL: redactURLUserInfo(c.QueueURL)}
	}
	return row
}

//// LIST FUNCTION

func listFunctionTriggers(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
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
		plugin.Logger(ctx).Error("scaleway_function_trigger.listFunctionTriggers", "connection_error", err)
		return nil, err
	}

	api := function.NewAPI(client)

	req := &function.ListTriggersRequest{
		Region:      parent.Region,
		NamespaceID: scw.StringPtr(parent.ID),
		Page:        scw.Int32Ptr(1),
		PageSize:    scw.Uint32Ptr(listPageSize(d)),
	}

	var count int

	for {
		resp, err := api.ListTriggers(req)
		if err != nil {
			plugin.Logger(ctx).Error("scaleway_function_trigger.listFunctionTriggers", "query_error", err)
			if is403Error(err) || isNotFoundOrUnsupportedZoneError(err) {
				return nil, nil
			}
			return nil, err
		}

		for _, item := range resp.Triggers {
			if item == nil {
				continue
			}
			if id := d.EqualsQualString("function_id"); id != "" && id != item.FunctionID {
				continue
			}
			d.StreamListItem(ctx, newFunctionTriggerRow(item, parent))
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
