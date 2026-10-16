package scaleway

import (
	"context"
	"testing"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

func TestNewTablesHaveUniqueColumns(t *testing.T) {
	ctx := context.Background()
	tables := []*plugin.Table{
		tableScalewayLB(ctx),
		tableScalewayIamApplication(ctx),
		tableScalewayIamGroup(ctx),
		tableScalewayIamPolicy(ctx),
		tableScalewayIamOrganizationSecuritySettings(ctx),
		tableScalewayAuditTrailAlertRule(ctx),
	}
	for _, tb := range tables {
		seen := map[string]bool{}
		for _, c := range tb.Columns {
			if seen[c.Name] {
				t.Errorf("%s: duplicate column %q", tb.Name, c.Name)
			}
			seen[c.Name] = true
		}
		if !seen["title"] || !seen["akas"] {
			t.Errorf("%s: missing title/akas", tb.Name)
		}
	}
}
