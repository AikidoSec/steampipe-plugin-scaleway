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
		tableScalewayBlockSnapshot(ctx),
		tableScalewayBlockVolume(ctx),
		tableScalewayKeyManagerKey(ctx),
		tableScalewayLBBackend(ctx),
		tableScalewayLBCertificate(ctx),
		tableScalewayLBFrontend(ctx),
		tableScalewayLBIP(ctx),
		tableScalewayRDBBackup(ctx),
		tableScalewayRDBPrivilege(ctx),
		tableScalewayRDBSnapshot(ctx),
		tableScalewayRDBUser(ctx),
		tableScalewayRegistryTag(ctx),
		tableScalewaySecret(ctx),
		tableScalewayEdgeServicesPipeline(ctx),
		tableScalewayFlexibleIP(ctx),
		tableScalewayInstancePlacementGroup(ctx),
		tableScalewayInstancePrivateNIC(ctx),
		tableScalewayInstanceTemplate(ctx),
		tableScalewayInterlinkDedicatedConnection(ctx),
		tableScalewayInterlinkLink(ctx),
		tableScalewayInterlinkRoutingPolicy(ctx),
		tableScalewayIPAMIP(ctx),
		tableScalewayS2SVPNConnection(ctx),
		tableScalewayS2SVPNCustomerGateway(ctx),
		tableScalewayS2SVPNGateway(ctx),
		tableScalewayS2SVPNRoutingPolicy(ctx),
		tableScalewayVPCConnector(ctx),
		tableScalewayVPCIngressRule(ctx),
		tableScalewayVPCSubnet(ctx),
		tableScalewayVPCGWGateway(ctx),
		tableScalewayVPCGWGatewayNetwork(ctx),
		tableScalewayVPCGWIP(ctx),
		tableScalewayVPCGWPATRule(ctx),
		tableScalewayContainer(ctx),
		tableScalewayContainerNamespace(ctx),
		tableScalewayContainerTrigger(ctx),
		tableScalewayDocumentDBInstance(ctx),
		tableScalewayDocumentDBSnapshot(ctx),
		tableScalewayFunction(ctx),
		tableScalewayFunctionNamespace(ctx),
		tableScalewayFunctionTrigger(ctx),
		tableScalewayJobDefinition(ctx),
		tableScalewayMongoDBInstance(ctx),
		tableScalewayMongoDBSnapshot(ctx),
		tableScalewayOpenSearchDeployment(ctx),
		tableScalewayRedisCluster(ctx),
		tableScalewayServerlessSQLDBBackup(ctx),
		tableScalewayServerlessSQLDBDatabase(ctx),
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
