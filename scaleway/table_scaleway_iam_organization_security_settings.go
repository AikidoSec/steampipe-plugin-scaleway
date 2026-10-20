package scaleway

import (
	"context"
	"errors"

	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

//// TABLE DEFINITION

func tableScalewayIamOrganizationSecuritySettings(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "scaleway_iam_organization_security_settings",
		Description: "Organization-wide IAM security settings such as API key expiration and login policies.",
		List: &plugin.ListConfig{
			Hydrate: listIamOrganizationSecuritySettings,
		},
		Columns: []*plugin.Column{
			{
				Name:        "organization_id",
				Description: "The ID of the organization the settings belong to.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("OrganizationID"),
			},
			{
				Name:        "max_api_key_expiration_duration",
				Description: "Maximum API key lifetime in seconds. 0 means no maximum.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("MaxAPIKeyExpirationDuration").Transform(durationToSeconds),
			},
			{
				Name:        "enforce_password_renewal",
				Description: "Whether periodic password renewal is enforced.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "login_attempts_before_locked",
				Description: "The number of failed login attempts before an account is locked.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("LoginAttemptsBeforeLocked"),
			},
			{
				Name:        "max_login_session_duration",
				Description: "Maximum login session lifetime in seconds.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("MaxLoginSessionDuration").Transform(durationToSeconds),
			},
			{
				Name:        "grace_period_duration",
				Description: "Grace period before password renewal is enforced, in seconds.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("GracePeriodDuration").Transform(durationToSeconds),
			},

			// Steampipe standard columns
			{
				Name:        "title",
				Description: "Title of the resource.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.From(iamOrganizationSecuritySettingsTitle),
			},
			{
				Name:        "akas",
				Description: "Array of globally unique identifier strings (also known as) for the resource.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("OrganizationID").Transform(transform.EnsureStringArray),
			},
		},
	}
}

// iamOrganizationSecuritySettingsRow flattens the settings together with the
// organization ID, which the API response does not carry.
type iamOrganizationSecuritySettingsRow struct {
	OrganizationID              string
	OrganizationName            string
	MaxAPIKeyExpirationDuration *scw.Duration
	EnforcePasswordRenewal      bool
	LoginAttemptsBeforeLocked   uint32
	MaxLoginSessionDuration     *scw.Duration
	GracePeriodDuration         *scw.Duration
}

//// TRANSFORM FUNCTIONS

// durationToSeconds renders a *scw.Duration as whole seconds (sub-second part dropped), or nil when unset.
func durationToSeconds(_ context.Context, d *transform.TransformData) (interface{}, error) {
	dur, ok := d.Value.(*scw.Duration)
	if !ok || dur == nil {
		return nil, nil
	}
	return dur.Seconds, nil
}

// iamOrganizationSecuritySettingsTitle uses the organization name, falling
// back to the organization ID when the name could not be read.
func iamOrganizationSecuritySettingsTitle(_ context.Context, d *transform.TransformData) (interface{}, error) {
	row, ok := d.HydrateItem.(iamOrganizationSecuritySettingsRow)
	if !ok {
		return nil, nil
	}
	if row.OrganizationName != "" {
		return row.OrganizationName, nil
	}
	return row.OrganizationID, nil
}

//// LIST FUNCTION

func listIamOrganizationSecuritySettings(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	logger := plugin.Logger(ctx)
	logger.Warn("scaleway_iam_organization_security_settings.listIamOrganizationSecuritySettings", "status", "start")

	organisationId := GetConfig(d.Connection).OrganizationID
	if organisationId == nil || *organisationId == "" {
		logger.Error("scaleway_iam_organization_security_settings.listIamOrganizationSecuritySettings", "config_error", "organization_id is not configured")
		return nil, errors.New("organization_id must be configured to query scaleway_iam_organization_security_settings")
	}

	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_iam_organization_security_settings.listIamOrganizationSecuritySettings", "connection_error", err)
		return nil, err
	}

	iamApi := iam.NewAPI(client)

	settings, err := iamApi.GetOrganizationSecuritySettings(&iam.GetOrganizationSecuritySettingsRequest{
		OrganizationID: *organisationId,
	})
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_iam_organization_security_settings.listIamOrganizationSecuritySettings", "query_error", err)
		return nil, err
	}
	logger.Warn("scaleway_iam_organization_security_settings.listIamOrganizationSecuritySettings", "status", "settings_fetched", "organization_id", *organisationId)

	// The name only feeds the title, so a missing permission leaves the ID as title.
	organizationName := ""
	organization, err := iamApi.GetOrganization(&iam.GetOrganizationRequest{OrganizationID: *organisationId})
	switch {
	case err == nil:
		organizationName = organization.Name
		logger.Warn("scaleway_iam_organization_security_settings.listIamOrganizationSecuritySettings", "status", "organization_fetched", "organization_name", organizationName)
	case is403Error(err) || is404Error(err):
		plugin.Logger(ctx).Warn("scaleway_iam_organization_security_settings.listIamOrganizationSecuritySettings", "organization_name_unavailable", err)
	default:
		plugin.Logger(ctx).Error("scaleway_iam_organization_security_settings.listIamOrganizationSecuritySettings", "organization_query_error", err)
		return nil, err
	}

	d.StreamListItem(ctx, iamOrganizationSecuritySettingsRow{
		OrganizationID:              *organisationId,
		OrganizationName:            organizationName,
		MaxAPIKeyExpirationDuration: settings.MaxAPIKeyExpirationDuration,
		EnforcePasswordRenewal:      settings.EnforcePasswordRenewal,
		LoginAttemptsBeforeLocked:   settings.LoginAttemptsBeforeLocked,
		MaxLoginSessionDuration:     settings.MaxLoginSessionDuration,
		GracePeriodDuration:         settings.GracePeriodDuration,
	})
	logger.Warn("scaleway_iam_organization_security_settings.listIamOrganizationSecuritySettings", "status", "row_streamed", "organization_id", *organisationId)

	return nil, nil
}
