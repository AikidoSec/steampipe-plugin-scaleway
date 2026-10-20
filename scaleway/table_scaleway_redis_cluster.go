package scaleway

import (
	"context"

	"github.com/scaleway/scaleway-sdk-go/api/redis/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

//// TABLE DEFINITION

func tableScalewayRedisCluster(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:              "scaleway_redis_cluster",
		Description:       "A Managed Database for Redis cluster. Credentials are not exposed.",
		GetMatrixItemFunc: BuildZoneList,
		List: &plugin.ListConfig{
			Hydrate: listRedisClusters,
			KeyColumns: []*plugin.KeyColumn{
				{
					Name:    "zone",
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
				Description: "The unique identifier of the cluster.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID"),
			},
			{
				Name:        "name",
				Description: "The name of the cluster.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "status",
				Description: "The status of the cluster.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Status").Transform(transform.ToString),
			},
			{
				Name:        "version",
				Description: "The Redis version of the cluster.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "node_type",
				Description: "The node type of the cluster.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "cluster_size",
				Description: "The number of nodes in the cluster.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("ClusterSize"),
			},
			{
				Name:        "tls_enabled",
				Description: "Indicates whether TLS is enabled.",
				Type:        proto.ColumnType_BOOL,
				Transform:   transform.FromField("TLSEnabled"),
			},
			{
				Name:        "user_name",
				Description: "The name of the initial user of the cluster.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "endpoints",
				Description: "The endpoints of the cluster.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "acl_rules",
				Description: "The ACL rules of the cluster.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("ACLRules"),
			},
			{
				Name:        "cluster_settings",
				Description: "The settings of the cluster.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "upgradable_versions",
				Description: "The versions the cluster can be upgraded to.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "srn",
				Description: "The Scaleway resource name of the cluster.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Srn"),
			},
			{
				Name:        "tags",
				Description: "A list of tags attached to the cluster.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "created_at",
				Description: "The time when the cluster was created.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "updated_at",
				Description: "The time when the cluster was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
			},

			// Scaleway standard columns
			{
				Name:        "region",
				Description: "The region where the cluster resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Zone").Transform(lbRegionFromZone),
			},
			{
				Name:        "zone",
				Description: "The zone where the cluster resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Zone").Transform(transform.ToString),
			},
			{
				Name:        "project_id",
				Description: "The ID of the project where the cluster resides.",
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

func listRedisClusters(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	zone, err := scw.ParseZone(d.EqualsQualString("zone"))
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_redis_cluster.listRedisClusters", "zone_parsing_error", err)
		return nil, err
	}

	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_redis_cluster.listRedisClusters", "connection_error", err)
		return nil, err
	}

	api := redis.NewAPI(client)

	req := &redis.ListClustersRequest{
		Zone:     zone,
		Page:     scw.Int32Ptr(1),
		PageSize: scw.Uint32Ptr(listPageSize(d)),
	}
	if project := d.EqualsQualString("project_id"); project != "" {
		req.ProjectID = scw.StringPtr(project)
	}

	var count int

	for {
		resp, err := api.ListClusters(req)
		if err != nil {
			plugin.Logger(ctx).Error("scaleway_redis_cluster.listRedisClusters", "query_error", err)
			if isNotFoundOrUnsupportedZoneError(err) || is403Error(err) {
				return nil, nil
			}
			return nil, err
		}

		for _, item := range resp.Clusters {
			if item == nil {
				continue
			}
			d.StreamListItem(ctx, item)
			count++

			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		if len(resp.Clusters) == 0 || uint64(count) >= uint64(resp.TotalCount) {
			break
		}
		req.Page = scw.Int32Ptr(*req.Page + 1)
	}

	return nil, nil
}
