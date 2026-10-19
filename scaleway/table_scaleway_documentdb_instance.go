package scaleway

import (
	"context"

	documentdb "github.com/scaleway/scaleway-sdk-go/api/documentdb/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

//// TABLE DEFINITION

func tableScalewayDocumentDBInstance(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:              "scaleway_documentdb_instance",
		Description:       "A Managed Database for DocumentDB instance. Credentials are not exposed.",
		GetMatrixItemFunc: BuildRegionList,
		List: &plugin.ListConfig{
			Hydrate: listDocumentDBInstances,
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
				Description: "The unique identifier of the instance.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID"),
			},
			{
				Name:        "name",
				Description: "The name of the instance.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "status",
				Description: "The status of the instance.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Status").Transform(transform.ToString),
			},
			{
				Name:        "engine",
				Description: "The database engine of the instance.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "node_type",
				Description: "The node type of the instance.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "is_ha_cluster",
				Description: "Indicates whether the instance is a high availability cluster.",
				Type:        proto.ColumnType_BOOL,
				Transform:   transform.FromField("IsHaCluster"),
			},
			{
				Name:        "backup_same_region",
				Description: "Indicates whether backups are stored in the same region.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "volume",
				Description: "The volume of the instance.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "endpoints",
				Description: "The endpoints of the instance.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "backup_schedule",
				Description: "The backup schedule of the instance.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "read_replicas",
				Description: "The read replicas of the instance.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "logs_policy",
				Description: "The logs policy of the instance.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "settings",
				Description: "The settings of the instance.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "init_settings",
				Description: "The initial settings of the instance.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "upgradable_version",
				Description: "The versions the instance can be upgraded to.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "maintenances",
				Description: "The maintenances of the instance.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "acl_rules",
				Description: "The ACL rules of the instance.",
				Type:        proto.ColumnType_JSON,
				Hydrate:     getDocumentDBInstanceACLRules,
				Transform:   transform.FromValue(),
			},
			{
				Name:        "tags",
				Description: "A list of tags attached to the instance.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "created_at",
				Description: "The time when the instance was created.",
				Type:        proto.ColumnType_TIMESTAMP,
			},

			// Scaleway standard columns
			{
				Name:        "region",
				Description: "The region where the instance resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Region").Transform(transform.ToString),
			},
			{
				Name:        "project_id",
				Description: "The ID of the project where the instance resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ProjectID"),
			},
			{
				Name:        "organization_id",
				Description: "The ID of the organization where the instance resides.",
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

func listDocumentDBInstances(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	region, err := scw.ParseRegion(d.EqualsQualString("region"))
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_documentdb_instance.listDocumentDBInstances", "region_parsing_error", err)
		return nil, err
	}

	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_documentdb_instance.listDocumentDBInstances", "connection_error", err)
		return nil, err
	}

	api := documentdb.NewAPI(client)

	req := &documentdb.ListInstancesRequest{
		Region:   region,
		Page:     scw.Int32Ptr(1),
		PageSize: scw.Uint32Ptr(listPageSize(d)),
	}
	if project := d.EqualsQualString("project_id"); project != "" {
		req.ProjectID = scw.StringPtr(project)
	}

	var count int

	for {
		resp, err := api.ListInstances(req)
		if err != nil {
			plugin.Logger(ctx).Error("scaleway_documentdb_instance.listDocumentDBInstances", "query_error", err)
			if is403Error(err) || isNotFoundOrUnsupportedZoneError(err) {
				return nil, nil
			}
			return nil, err
		}

		for _, item := range resp.Instances {
			if item == nil {
				continue
			}
			d.StreamListItem(ctx, item)
			count++

			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		if len(resp.Instances) == 0 || uint64(count) >= uint64(resp.TotalCount) {
			break
		}
		req.Page = scw.Int32Ptr(*req.Page + 1)
	}

	return nil, nil
}

//// HYDRATE FUNCTIONS

func getDocumentDBInstanceACLRules(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	instance := h.Item.(*documentdb.Instance)

	if instance.Region.String() != d.EqualsQualString("region") {
		return nil, nil
	}

	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_documentdb_instance.getDocumentDBInstanceACLRules", "connection_error", err)
		return nil, err
	}

	api := documentdb.NewAPI(client)

	req := &documentdb.ListInstanceACLRulesRequest{
		Region:     instance.Region,
		InstanceID: instance.ID,
		Page:       scw.Int32Ptr(1),
		PageSize:   scw.Uint32Ptr(maxListPageSize),
	}

	rules := []*documentdb.ACLRule{}
	for {
		resp, err := api.ListInstanceACLRules(req)
		if err != nil {
			if is403Error(err) || isNotFoundOrUnsupportedZoneError(err) {
				return nil, nil
			}
			plugin.Logger(ctx).Error("scaleway_documentdb_instance.getDocumentDBInstanceACLRules", "query_error", err)
			return nil, err
		}

		rules = append(rules, resp.Rules...)

		if len(resp.Rules) == 0 || uint32(len(rules)) >= resp.TotalCount {
			break
		}
		req.Page = scw.Int32Ptr(*req.Page + 1)
	}

	return rules, nil
}
