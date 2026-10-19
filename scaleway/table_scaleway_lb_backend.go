package scaleway

import (
	"context"

	lb "github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

//// TABLE DEFINITION

func tableScalewayLBBackend(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:              "scaleway_lb_backend",
		Description:       "A load balancer backend is a pool of servers that receives the traffic forwarded by a frontend.",
		GetMatrixItemFunc: BuildZoneList,
		List: &plugin.ListConfig{
			Hydrate:       listLBBackends,
			ParentHydrate: listLBs,
			KeyColumns: []*plugin.KeyColumn{
				{
					Name:    "zone",
					Require: plugin.Optional,
				},
				{
					Name:    "lb_id",
					Require: plugin.Optional,
				},
			},
		},
		Columns: []*plugin.Column{
			{
				Name:        "id",
				Description: "The unique identifier of the backend.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID"),
			},
			{
				Name:        "name",
				Description: "The name of the backend.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "lb_id",
				Description: "The ID of the load balancer the backend belongs to.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("LBID"),
			},
			{
				Name:        "forward_protocol",
				Description: "The protocol used to forward traffic to the backend servers.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ForwardProtocol").Transform(transform.ToString),
			},
			{
				Name:        "forward_port",
				Description: "The port used to forward traffic to the backend servers.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "forward_port_algorithm",
				Description: "The load balancing algorithm.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ForwardPortAlgorithm").Transform(transform.ToString),
			},
			{
				Name:        "sticky_sessions",
				Description: "The type of sticky sessions.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("StickySessions").Transform(transform.ToString),
			},
			{
				Name:        "sticky_sessions_cookie_name",
				Description: "The cookie name used for sticky sessions.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "health_check",
				Description: "The health check configuration of the backend.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "pool",
				Description: "The IP addresses of the backend servers.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "send_proxy_v2",
				Description: "Indicates whether PROXY protocol version 2 is sent to the backend servers (deprecated by proxy_protocol).",
				Type:        proto.ColumnType_BOOL,
				Transform:   transform.FromField("SendProxyV2"),
			},
			{
				Name:        "proxy_protocol",
				Description: "The PROXY protocol version sent to the backend servers.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ProxyProtocol").Transform(transform.ToString),
			},
			{
				Name:        "ssl_bridging",
				Description: "Indicates whether SSL is used between the load balancer and the backend servers.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "ignore_ssl_server_verify",
				Description: "Indicates whether the verification of the backend servers' certificates is skipped.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "on_marked_down_action",
				Description: "The action taken when a backend server is marked down.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("OnMarkedDownAction").Transform(transform.ToString),
			},
			{
				Name:        "failover_host",
				Description: "The host serving the maintenance page when all backend servers are down.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "max_connections",
				Description: "The maximum number of connections allowed per backend server.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "max_retries",
				Description: "The number of retries when connecting to a backend server.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "redispatch_attempt_count",
				Description: "The number of redispatches on connection failure.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "timeout_server",
				Description: "The maximum inactivity time on the backend server side, in nanoseconds.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "timeout_connect",
				Description: "The maximum time to establish a connection to a backend server, in nanoseconds.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "timeout_tunnel",
				Description: "The maximum tunnel inactivity time, in nanoseconds.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "created_at",
				Description: "The time when the backend was created.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "updated_at",
				Description: "The time when the backend was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
			},

			// Scaleway standard columns
			{
				Name:        "region",
				Description: "The region where the backend resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Zone").Transform(lbRegionFromZone),
			},
			{
				Name:        "zone",
				Description: "The zone where the backend resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Zone").Transform(transform.ToString),
			},
			{
				Name:        "project_id",
				Description: "The ID of the project where the backend resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ProjectID"),
			},
			{
				Name:        "organization_id",
				Description: "The ID of the organization where the backend resides.",
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

type lbBackendRow struct {
	lb.Backend
	LBID           string
	Zone           scw.Zone
	ProjectID      string
	OrganizationID string
}

//// LIST FUNCTION

func listLBBackends(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	matrixZone := d.EqualsQualString("zone")
	parent := h.Item.(*lb.LB)

	if parent.Zone.String() != matrixZone {
		return nil, nil
	}
	if id := d.EqualsQualString("lb_id"); id != "" && id != parent.ID {
		return nil, nil
	}

	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_lb_backend.listLBBackends", "connection_error", err)
		return nil, err
	}

	lbApi := lb.NewZonedAPI(client)

	req := &lb.ZonedAPIListBackendsRequest{
		Zone:     parent.Zone,
		LBID:     parent.ID,
		Page:     scw.Int32Ptr(1),
		PageSize: scw.Uint32Ptr(listPageSize(d)),
	}

	var count int

	for {
		resp, err := lbApi.ListBackends(req)
		if err != nil {
			plugin.Logger(ctx).Error("scaleway_lb_backend.listLBBackends", "query_error", err)
			if isNotFoundOrUnsupportedZoneError(err) {
				return nil, nil
			}
			return nil, err
		}

		for _, backend := range resp.Backends {
			d.StreamListItem(ctx, lbBackendRow{*backend, parent.ID, parent.Zone, parent.ProjectID, parent.OrganizationID})
			count++

			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		if len(resp.Backends) == 0 || resp.TotalCount == uint32(count) {
			break
		}
		req.Page = scw.Int32Ptr(*req.Page + 1)
	}

	return nil, nil
}
