package scaleway

import (
	"context"

	audit_trail "github.com/scaleway/scaleway-sdk-go/api/audit_trail/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

//// TABLE DEFINITION

func tableScalewayAuditTrailAlertRule(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:              "scaleway_audit_trail_alert_rule",
		Description:       "Audit Trail alert rules notify on security-relevant events recorded in the audit trail.",
		GetMatrixItemFunc: BuildRegionList,
		List: &plugin.ListConfig{
			Hydrate: listAuditTrailAlertRules,
			KeyColumns: []*plugin.KeyColumn{
				{
					Name:    "region",
					Require: plugin.Optional,
				},
			},
		},
		Columns: []*plugin.Column{
			{
				Name:        "id",
				Description: "The unique identifier of the alert rule.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID"),
			},
			{
				Name:        "name",
				Description: "The name of the alert rule.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "description",
				Description: "The description of the alert rule.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "status",
				Description: "The status of the alert rule.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Status").Transform(transform.ToString),
			},

			// Scaleway standard columns
			{
				Name:        "region",
				Description: "The region where the alert rule resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromMatrixItem("region"),
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

func listAuditTrailAlertRules(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	region := d.EqualsQualString("region")

	parseRegionData, err := scw.ParseRegion(region)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_audit_trail_alert_rule.listAuditTrailAlertRules", "region_parsing_error", err)
		return nil, err
	}

	quals := d.EqualsQuals
	if quals["region"] != nil && quals["region"].GetStringValue() != region {
		return nil, nil
	}

	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_audit_trail_alert_rule.listAuditTrailAlertRules", "connection_error", err)
		return nil, err
	}

	auditApi := audit_trail.NewAPI(client)

	req := &audit_trail.ListAlertRulesRequest{
		Region: parseRegionData,
		Page:   scw.Int32Ptr(1),
	}
	if organisationId := GetConfig(d.Connection).OrganizationID; organisationId != nil {
		req.OrganizationID = *organisationId
	}

	maxResult := int64(100)

	// Reduce the basic request limit down if the user has only requested a small number of rows
	limit := d.QueryContext.Limit
	if d.QueryContext.Limit != nil {
		if *limit < maxResult {
			maxResult = *limit
		}
	}
	req.PageSize = scw.Uint32Ptr(uint32(maxResult))

	var count int

	for {
		resp, err := auditApi.ListAlertRules(req)
		if err != nil {
			plugin.Logger(ctx).Error("scaleway_audit_trail_alert_rule.listAuditTrailAlertRules", "query_error", err)
			return nil, err
		}

		for _, rule := range resp.AlertRules {
			d.StreamListItem(ctx, rule)

			// Increase the resource count by 1
			count++

			// Context can be cancelled due to manual cancellation or the limit has been hit
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		if uint64(count) >= resp.TotalCount || len(resp.AlertRules) == 0 {
			break
		}
		req.Page = scw.Int32Ptr(*req.Page + 1)
	}

	return nil, nil
}
