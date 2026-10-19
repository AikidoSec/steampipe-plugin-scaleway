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

func tableScalewayLBFrontend(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:              "scaleway_lb_frontend",
		Description:       "A load balancer frontend listens on a port and forwards the incoming traffic to a backend.",
		GetMatrixItemFunc: BuildZoneList,
		List: &plugin.ListConfig{
			Hydrate:       listLBFrontends,
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
				Description: "The unique identifier of the frontend.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID"),
			},
			{
				Name:        "name",
				Description: "The name of the frontend.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "lb_id",
				Description: "The ID of the load balancer the frontend belongs to.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("LBID"),
			},
			{
				Name:        "backend_id",
				Description: "The ID of the backend the frontend forwards traffic to.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("BackendID"),
			},
			{
				Name:        "inbound_port",
				Description: "The port the frontend listens on.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "certificate_ids",
				Description: "The IDs of the certificates attached to the frontend.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("CertificateIDs"),
			},
			{
				Name:        "enable_http3",
				Description: "Indicates whether HTTP/3 is enabled on the frontend.",
				Type:        proto.ColumnType_BOOL,
				Transform:   transform.FromField("EnableHTTP3"),
			},
			{
				Name:        "enable_access_logs",
				Description: "Indicates whether access logs are enabled on the frontend.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "connection_rate_limit",
				Description: "The maximum number of new connections per second.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "timeout_client",
				Description: "The maximum inactivity time on the client side, in nanoseconds.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "acls",
				Description: "The ACLs of the frontend.",
				Type:        proto.ColumnType_JSON,
				Hydrate:     getLBFrontendACLs,
				Transform:   transform.FromValue(),
			},
			{
				Name:        "routes",
				Description: "The routes of the frontend.",
				Type:        proto.ColumnType_JSON,
				Hydrate:     getLBFrontendRoutes,
				Transform:   transform.FromValue(),
			},
			{
				Name:        "created_at",
				Description: "The time when the frontend was created.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "updated_at",
				Description: "The time when the frontend was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
			},

			// Scaleway standard columns
			{
				Name:        "region",
				Description: "The region where the frontend resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Zone").Transform(lbRegionFromZone),
			},
			{
				Name:        "zone",
				Description: "The zone where the frontend resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Zone").Transform(transform.ToString),
			},
			{
				Name:        "project_id",
				Description: "The ID of the project where the frontend resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ProjectID"),
			},
			{
				Name:        "organization_id",
				Description: "The ID of the organization where the frontend resides.",
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

type lbFrontendRow struct {
	lb.Frontend
	LBID           string
	BackendID      string
	Zone           scw.Zone
	ProjectID      string
	OrganizationID string
}

//// LIST FUNCTION

func listLBFrontends(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
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
		plugin.Logger(ctx).Error("scaleway_lb_frontend.listLBFrontends", "connection_error", err)
		return nil, err
	}

	lbApi := lb.NewZonedAPI(client)

	req := &lb.ZonedAPIListFrontendsRequest{
		Zone:     parent.Zone,
		LBID:     parent.ID,
		Page:     scw.Int32Ptr(1),
		PageSize: scw.Uint32Ptr(listPageSize(d)),
	}

	var count int

	for {
		resp, err := lbApi.ListFrontends(req)
		if err != nil {
			plugin.Logger(ctx).Error("scaleway_lb_frontend.listLBFrontends", "query_error", err)
			if isNotFoundOrUnsupportedZoneError(err) {
				return nil, nil
			}
			return nil, err
		}

		for _, frontend := range resp.Frontends {
			var backendID string
			if frontend.Backend != nil {
				backendID = frontend.Backend.ID
			}
			d.StreamListItem(ctx, lbFrontendRow{*frontend, parent.ID, backendID, parent.Zone, parent.ProjectID, parent.OrganizationID})
			count++

			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		if len(resp.Frontends) == 0 || resp.TotalCount == uint32(count) {
			break
		}
		req.Page = scw.Int32Ptr(*req.Page + 1)
	}

	return nil, nil
}

//// HYDRATE FUNCTIONS

func getLBFrontendACLs(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	matrixZone := d.EqualsQualString("zone")
	frontend := h.Item.(lbFrontendRow)

	if matrixZone != frontend.Zone.String() {
		return nil, nil
	}

	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_lb_frontend.getLBFrontendACLs", "connection_error", err)
		return nil, err
	}

	lbApi := lb.NewZonedAPI(client)

	req := &lb.ZonedAPIListACLsRequest{
		Zone:       frontend.Zone,
		FrontendID: frontend.ID,
		Page:       scw.Int32Ptr(1),
		PageSize:   scw.Uint32Ptr(maxListPageSize),
	}

	acls := []*lb.ACL{}
	for {
		resp, err := lbApi.ListACLs(req)
		if err != nil {
			if is403Error(err) || isNotFoundOrUnsupportedZoneError(err) {
				return nil, nil
			}
			plugin.Logger(ctx).Error("scaleway_lb_frontend.getLBFrontendACLs", "query_error", err)
			return nil, err
		}

		acls = append(acls, resp.ACLs...)

		if len(resp.ACLs) == 0 || uint32(len(acls)) >= resp.TotalCount {
			break
		}
		req.Page = scw.Int32Ptr(*req.Page + 1)
	}

	return acls, nil
}

func getLBFrontendRoutes(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	matrixZone := d.EqualsQualString("zone")
	frontend := h.Item.(lbFrontendRow)

	if matrixZone != frontend.Zone.String() {
		return nil, nil
	}

	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_lb_frontend.getLBFrontendRoutes", "connection_error", err)
		return nil, err
	}

	lbApi := lb.NewZonedAPI(client)

	req := &lb.ZonedAPIListRoutesRequest{
		Zone:       frontend.Zone,
		FrontendID: scw.StringPtr(frontend.ID),
		Page:       scw.Int32Ptr(1),
		PageSize:   scw.Uint32Ptr(maxListPageSize),
	}

	routes := []*lb.Route{}
	for {
		resp, err := lbApi.ListRoutes(req)
		if err != nil {
			if is403Error(err) || isNotFoundOrUnsupportedZoneError(err) {
				return nil, nil
			}
			plugin.Logger(ctx).Error("scaleway_lb_frontend.getLBFrontendRoutes", "query_error", err)
			return nil, err
		}

		routes = append(routes, resp.Routes...)

		if len(resp.Routes) == 0 || uint32(len(routes)) >= resp.TotalCount {
			break
		}
		req.Page = scw.Int32Ptr(*req.Page + 1)
	}

	return routes, nil
}
