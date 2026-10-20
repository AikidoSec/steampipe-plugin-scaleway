package scaleway

import (
	"context"

	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

//// TABLE DEFINITION

func tableScalewayIamPolicy(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "scaleway_iam_policy",
		Description: "Policies grant permissions to a user, group or application through a set of rules.",
		List: &plugin.ListConfig{
			Hydrate: listIamPolicies,
		},
		Get: &plugin.GetConfig{
			Hydrate:    getIamPolicy,
			KeyColumns: plugin.SingleColumn("id"),
		},
		Columns: []*plugin.Column{
			{
				Name:        "id",
				Description: "The unique identifier of the policy.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID"),
			},
			{
				Name:        "name",
				Description: "The name of the policy.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "description",
				Description: "The description of the policy.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "editable",
				Description: "Whether the policy can be edited.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "managed",
				Description: "Whether the policy is managed by Scaleway.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "nb_rules",
				Description: "The number of rules in the policy.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("NbRules"),
			},
			{
				Name:        "nb_scopes",
				Description: "The number of scopes in the policy.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("NbScopes"),
			},
			{
				Name:        "nb_permission_sets",
				Description: "The number of permission sets in the policy.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("NbPermissionSets"),
			},
			{
				Name:        "user_id",
				Description: "The ID of the user the policy is attached to, if any.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("UserID"),
			},
			{
				Name:        "group_id",
				Description: "The ID of the group the policy is attached to, if any.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("GroupID"),
			},
			{
				Name:        "application_id",
				Description: "The ID of the application the policy is attached to, if any.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ApplicationID"),
			},
			{
				Name:        "no_principal",
				Description: "True when the policy is not attached to any principal.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "tags",
				Description: "A list of tags attached to the policy.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "rules",
				Description: "The rules of the policy: permission set names, scope type, condition, project IDs and organization ID.",
				Type:        proto.ColumnType_JSON,
				Hydrate:     getIamPolicyRules,
				Transform:   transform.FromValue().Transform(iamRulesToJSON),
			},
			{
				Name:        "created_at",
				Description: "The time when the policy was created.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "updated_at",
				Description: "The time when the policy was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
			},

			// Scaleway standard columns
			{
				Name:        "organization",
				Description: "The ID of the organization where the policy resides.",
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

//// TRANSFORM FUNCTIONS

// iamRuleJSON is the serialized shape of an IAM rule. The SDK's *[]string
// fields become a nil slice (JSON null) when unset and an empty slice (JSON [])
// when set but empty.
type iamRuleJSON struct {
	ID                      string   `json:"id"`
	PermissionSetNames      []string `json:"permission_set_names"`
	PermissionSetsScopeType string   `json:"permission_sets_scope_type"`
	Condition               string   `json:"condition"`
	ProjectIDs              []string `json:"project_ids"`
	OrganizationID          *string  `json:"organization_id"`
}

func derefStringSlice(p *[]string) []string {
	if p == nil {
		return nil
	}
	if *p == nil {
		return []string{}
	}
	return *p
}

// iamRulesToJSON converts SDK rules into their serialized shape, or nil when
// the value is not a rule list.
func iamRulesToJSON(_ context.Context, d *transform.TransformData) (interface{}, error) {
	rules, ok := d.Value.([]*iam.Rule)
	if !ok {
		return nil, nil
	}
	out := make([]iamRuleJSON, 0, len(rules))
	for _, r := range rules {
		if r == nil {
			continue
		}
		out = append(out, iamRuleJSON{
			ID:                      r.ID,
			PermissionSetNames:      derefStringSlice(r.PermissionSetNames),
			PermissionSetsScopeType: r.PermissionSetsScopeType.String(),
			Condition:               r.Condition,
			ProjectIDs:              derefStringSlice(r.ProjectIDs),
			OrganizationID:          r.OrganizationID,
		})
	}
	return out, nil
}

//// LIST FUNCTION

func listIamPolicies(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_iam_policy.listIamPolicies", "connection_error", err)
		return nil, err
	}

	iamApi := iam.NewAPI(client)

	req := &iam.ListPoliciesRequest{
		Page: scw.Int32Ptr(1),
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
		resp, err := iamApi.ListPolicies(req)
		if err != nil {
			plugin.Logger(ctx).Error("scaleway_iam_policy.listIamPolicies", "query_error", err)
			return nil, err
		}

		for _, policy := range resp.Policies {
			d.StreamListItem(ctx, policy)

			// Increase the resource count by 1
			count++

			// Context can be cancelled due to manual cancellation or the limit has been hit
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		if resp.TotalCount == uint32(count) {
			break
		}
		req.Page = scw.Int32Ptr(*req.Page + 1)
	}

	return nil, nil
}

//// HYDRATE FUNCTIONS

func getIamPolicy(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	id := d.EqualsQualString("id")
	if id == "" {
		return nil, nil
	}

	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_iam_policy.getIamPolicy", "connection_error", err)
		return nil, err
	}

	iamApi := iam.NewAPI(client)

	data, err := iamApi.GetPolicy(&iam.GetPolicyRequest{
		PolicyID: id,
	})
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_iam_policy.getIamPolicy", "query_error", err)
		if is404Error(err) {
			return nil, nil
		}
		return nil, err
	}

	return data, nil
}

func getIamPolicyRules(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	policy := h.Item.(*iam.Policy)

	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_iam_policy.getIamPolicyRules", "connection_error", err)
		return nil, err
	}

	iamApi := iam.NewAPI(client)

	req := &iam.ListRulesRequest{
		PolicyID: policy.ID,
		Page:     scw.Int32Ptr(1),
		PageSize: scw.Uint32Ptr(100),
	}

	rules := []*iam.Rule{}
	for {
		resp, err := iamApi.ListRules(req)
		if err != nil {
			if is403Error(err) || is404Error(err) {
				return nil, nil
			}
			plugin.Logger(ctx).Error("scaleway_iam_policy.getIamPolicyRules", "query_error", err)
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
