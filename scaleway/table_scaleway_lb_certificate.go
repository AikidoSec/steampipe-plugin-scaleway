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

func tableScalewayLBCertificate(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:              "scaleway_lb_certificate",
		Description:       "A load balancer certificate is the TLS certificate a frontend uses to terminate HTTPS traffic.",
		GetMatrixItemFunc: BuildZoneList,
		List: &plugin.ListConfig{
			Hydrate:       listLBCertificates,
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
				Description: "The unique identifier of the certificate.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID"),
			},
			{
				Name:        "name",
				Description: "The name of the certificate.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "lb_id",
				Description: "The ID of the load balancer the certificate belongs to.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("LBID"),
			},
			{
				Name:        "type",
				Description: "The type of the certificate (Let's Encrypt or custom).",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Type").Transform(transform.ToString),
			},
			{
				Name:        "status",
				Description: "The status of the certificate.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Status").Transform(transform.ToString),
			},
			{
				Name:        "status_details",
				Description: "Additional details about the status of the certificate.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "common_name",
				Description: "The main domain name of the certificate.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "subject_alternative_name",
				Description: "The alternative domain names of the certificate.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "fingerprint",
				Description: "The SHA-1 fingerprint of the certificate.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "not_valid_before",
				Description: "The start of the validity period of the certificate.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "not_valid_after",
				Description: "The expiry time of the certificate.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "created_at",
				Description: "The time when the certificate was created.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "updated_at",
				Description: "The time when the certificate was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
			},

			// Scaleway standard columns
			{
				Name:        "region",
				Description: "The region where the certificate resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Zone").Transform(lbRegionFromZone),
			},
			{
				Name:        "zone",
				Description: "The zone where the certificate resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Zone").Transform(transform.ToString),
			},
			{
				Name:        "project_id",
				Description: "The ID of the project where the certificate resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ProjectID"),
			},
			{
				Name:        "organization_id",
				Description: "The ID of the organization where the certificate resides.",
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

type lbCertificateRow struct {
	lb.Certificate
	LBID           string
	Zone           scw.Zone
	ProjectID      string
	OrganizationID string
}

//// LIST FUNCTION

func listLBCertificates(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
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
		plugin.Logger(ctx).Error("scaleway_lb_certificate.listLBCertificates", "connection_error", err)
		return nil, err
	}

	lbApi := lb.NewZonedAPI(client)

	req := &lb.ZonedAPIListCertificatesRequest{
		Zone:     parent.Zone,
		LBID:     parent.ID,
		Page:     scw.Int32Ptr(1),
		PageSize: scw.Uint32Ptr(listPageSize(d)),
	}

	var count int

	for {
		resp, err := lbApi.ListCertificates(req)
		if err != nil {
			plugin.Logger(ctx).Error("scaleway_lb_certificate.listLBCertificates", "query_error", err)
			if isNotFoundOrUnsupportedZoneError(err) {
				return nil, nil
			}
			return nil, err
		}

		for _, certificate := range resp.Certificates {
			d.StreamListItem(ctx, lbCertificateRow{*certificate, parent.ID, parent.Zone, parent.ProjectID, parent.OrganizationID})
			count++

			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		if len(resp.Certificates) == 0 || resp.TotalCount == uint32(count) {
			break
		}
		req.Page = scw.Int32Ptr(*req.Page + 1)
	}

	return nil, nil
}
