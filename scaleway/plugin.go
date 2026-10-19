package scaleway

import (
	"context"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

const pluginName = "steampipe-plugin-scaleway"

// Plugin creates this (scaleway) plugin
func Plugin(ctx context.Context) *plugin.Plugin {
	p := &plugin.Plugin{
		Name:             pluginName,
		DefaultTransform: transform.FromCamel().Transform(transform.NullIfZeroValue),
		DefaultGetConfig: &plugin.GetConfig{},
		DefaultIgnoreConfig: &plugin.IgnoreConfig{
			ShouldIgnoreErrorFunc: shouldIgnoreErrorPluginDefault(),
		},
		ConnectionConfigSchema: &plugin.ConnectionConfigSchema{
			NewInstance: ConfigInstance,
		},
		TableMap: map[string]*plugin.Table{
			"scaleway_account_project":                    tableScalewayAccountProject(ctx),
			"scaleway_account_ssh_key":                    tableScalewayAccountSSHKey(ctx),
			"scaleway_audit_trail_alert_rule":             tableScalewayAuditTrailAlertRule(ctx),
			"scaleway_baremetal_server":                   tableScalewayBaremetalServer(ctx),
			"scaleway_billing_consumption":                tableScalewayBillingConsumption(ctx),
			"scaleway_billing_invoice":                    tableScalewayBillingInvoice(ctx),
			"scaleway_block_snapshot":                     tableScalewayBlockSnapshot(ctx),
			"scaleway_block_volume":                       tableScalewayBlockVolume(ctx),
			"scaleway_container":                          tableScalewayContainer(ctx),
			"scaleway_container_namespace":                tableScalewayContainerNamespace(ctx),
			"scaleway_container_trigger":                  tableScalewayContainerTrigger(ctx),
			"scaleway_documentdb_instance":                tableScalewayDocumentDBInstance(ctx),
			"scaleway_documentdb_snapshot":                tableScalewayDocumentDBSnapshot(ctx),
			"scaleway_edge_services_pipeline":             tableScalewayEdgeServicesPipeline(ctx),
			"scaleway_flexible_ip":                        tableScalewayFlexibleIP(ctx),
			"scaleway_function":                           tableScalewayFunction(ctx),
			"scaleway_function_namespace":                 tableScalewayFunctionNamespace(ctx),
			"scaleway_function_trigger":                   tableScalewayFunctionTrigger(ctx),
			"scaleway_iam_api_key":                        tableScalewayIamAPIKey(ctx),
			"scaleway_iam_application":                    tableScalewayIamApplication(ctx),
			"scaleway_iam_group":                          tableScalewayIamGroup(ctx),
			"scaleway_iam_organization_security_settings": tableScalewayIamOrganizationSecuritySettings(ctx),
			"scaleway_iam_policy":                         tableScalewayIamPolicy(ctx),
			"scaleway_iam_user":                           tableScalewayIamUser(ctx),
			"scaleway_instance_image":                     tableScalewayInstanceImage(ctx),
			"scaleway_instance_ip":                        tableScalewayInstanceIP(ctx),
			"scaleway_instance_placement_group":           tableScalewayInstancePlacementGroup(ctx),
			"scaleway_instance_private_nic":               tableScalewayInstancePrivateNIC(ctx),
			"scaleway_instance_security_group":            tableScalewayInstanceSecurityGroup(ctx),
			"scaleway_instance_server":                    tableScalewayInstanceServer(ctx),
			"scaleway_instance_snapshot":                  tableScalewayInstanceSnapshot(ctx),
			"scaleway_instance_template":                  tableScalewayInstanceTemplate(ctx),
			"scaleway_instance_volume":                    tableScalewayInstanceVolume(ctx),
			"scaleway_interlink_dedicated_connection":     tableScalewayInterlinkDedicatedConnection(ctx),
			"scaleway_interlink_link":                     tableScalewayInterlinkLink(ctx),
			"scaleway_interlink_routing_policy":           tableScalewayInterlinkRoutingPolicy(ctx),
			"scaleway_ipam_ip":                            tableScalewayIPAMIP(ctx),
			"scaleway_job_definition":                     tableScalewayJobDefinition(ctx),
			"scaleway_key_manager_key":                    tableScalewayKeyManagerKey(ctx),
			"scaleway_kubernetes_cluster":                 tableScalewayKubernetesCluster(ctx),
			"scaleway_kubernetes_node":                    tableScalewayKubernetesNode(ctx),
			"scaleway_kubernetes_pool":                    tableScalewayKubernetesPool(ctx),
			"scaleway_lb":                                 tableScalewayLB(ctx),
			"scaleway_lb_backend":                         tableScalewayLBBackend(ctx),
			"scaleway_lb_certificate":                     tableScalewayLBCertificate(ctx),
			"scaleway_lb_frontend":                        tableScalewayLBFrontend(ctx),
			"scaleway_lb_ip":                              tableScalewayLBIP(ctx),
			"scaleway_mongodb_instance":                   tableScalewayMongoDBInstance(ctx),
			"scaleway_mongodb_snapshot":                   tableScalewayMongoDBSnapshot(ctx),
			"scaleway_object_bucket":                      tableScalewayObjectBucket(ctx),
			"scaleway_opensearch_deployment":              tableScalewayOpenSearchDeployment(ctx),
			"scaleway_rdb_backup":                         tableScalewayRDBBackup(ctx),
			"scaleway_rdb_database":                       tableScalewayRDBDatabase(ctx),
			"scaleway_rdb_instance":                       tableScalewayRDBInstance(ctx),
			"scaleway_rdb_privilege":                      tableScalewayRDBPrivilege(ctx),
			"scaleway_rdb_snapshot":                       tableScalewayRDBSnapshot(ctx),
			"scaleway_rdb_user":                           tableScalewayRDBUser(ctx),
			"scaleway_redis_cluster":                      tableScalewayRedisCluster(ctx),
			"scaleway_registry_image":                     tableScalewayRegistryImage(ctx),
			"scaleway_registry_namespace":                 tableScalewayRegistryNamespace(ctx),
			"scaleway_registry_tag":                       tableScalewayRegistryTag(ctx),
			"scaleway_s2s_vpn_connection":                 tableScalewayS2SVPNConnection(ctx),
			"scaleway_s2s_vpn_customer_gateway":           tableScalewayS2SVPNCustomerGateway(ctx),
			"scaleway_s2s_vpn_gateway":                    tableScalewayS2SVPNGateway(ctx),
			"scaleway_s2s_vpn_routing_policy":             tableScalewayS2SVPNRoutingPolicy(ctx),
			"scaleway_secret":                             tableScalewaySecret(ctx),
			"scaleway_serverless_sqldb_backup":            tableScalewayServerlessSQLDBBackup(ctx),
			"scaleway_serverless_sqldb_database":          tableScalewayServerlessSQLDBDatabase(ctx),
			"scaleway_vpc":                                tableScalewayVPC(ctx),
			"scaleway_vpc_connector":                      tableScalewayVPCConnector(ctx),
			"scaleway_vpc_ingress_rule":                   tableScalewayVPCIngressRule(ctx),
			"scaleway_vpc_private_network":                tableScalewayVPCPrivateNetwork(ctx),
			"scaleway_vpc_subnet":                         tableScalewayVPCSubnet(ctx),
			"scaleway_vpcgw_gateway":                      tableScalewayVPCGWGateway(ctx),
			"scaleway_vpcgw_gateway_network":              tableScalewayVPCGWGatewayNetwork(ctx),
			"scaleway_vpcgw_ip":                           tableScalewayVPCGWIP(ctx),
			"scaleway_vpcgw_pat_rule":                     tableScalewayVPCGWPATRule(ctx),
		},
	}

	return p
}
