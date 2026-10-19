package scaleway

import (
	"context"

	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

//// TABLE DEFINITION

func tableScalewayInstancePrivateNIC(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:              "scaleway_instance_private_nic",
		Description:       "An Instance private NIC attaches a server to a private network.",
		GetMatrixItemFunc: BuildZoneList,
		List: &plugin.ListConfig{
			Hydrate:       listInstancePrivateNICs,
			ParentHydrate: listInstanceServers,
			KeyColumns: []*plugin.KeyColumn{
				{
					Name:    "zone",
					Require: plugin.Optional,
				},
				{
					Name:    "server_id",
					Require: plugin.Optional,
				},
			},
		},
		Columns: []*plugin.Column{
			{
				Name:        "id",
				Description: "The unique identifier of the private NIC.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID"),
			},
			{
				Name:        "server_id",
				Description: "The ID of the server the NIC is attached to.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ServerID"),
			},
			{
				Name:        "private_network_id",
				Description: "The ID of the private network the NIC is attached to.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("PrivateNetworkID"),
			},
			{
				Name:        "mac_address",
				Description: "The MAC address of the NIC.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("MacAddress"),
			},
			{
				Name:        "state",
				Description: "The state of the NIC.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("State").Transform(transform.ToString),
			},
			{
				Name:        "ipam_ip_ids",
				Description: "The IDs of the IPAM IP addresses of the NIC.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("IpamIPIDs"),
			},
			{
				Name:        "tags",
				Description: "A list of tags attached to the private NIC.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Tags"),
			},
			{
				Name:        "srn",
				Description: "The Scaleway resource name of the private NIC.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Srn"),
			},
			{
				Name:        "creation_date",
				Description: "The time when the NIC was created.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("CreationDate"),
			},

			// Scaleway standard columns
			{
				Name:        "region",
				Description: "The region where the private NIC resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Zone").Transform(lbRegionFromZone),
			},
			{
				Name:        "zone",
				Description: "The zone where the private NIC resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Zone").Transform(transform.ToString),
			},
			{
				Name:        "project_id",
				Description: "The ID of the project where the private NIC resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ProjectID"),
			},
			{
				Name:        "organization_id",
				Description: "The ID of the organization where the private NIC resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("OrganizationID"),
			},

			// Steampipe standard columns
			{
				Name:        "title",
				Description: "Title of the resource.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID"),
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

type instancePrivateNICRow struct {
	instance.PrivateNIC
	ProjectID      string
	OrganizationID string
}

//// LIST FUNCTION

func listInstancePrivateNICs(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	matrixLoc := d.EqualsQualString("zone")
	parent := h.Item.(*instance.Server)

	if parent.Zone.String() != matrixLoc {
		return nil, nil
	}
	if id := d.EqualsQualString("server_id"); id != "" && id != parent.ID {
		return nil, nil
	}

	for _, item := range parent.PrivateNics {
		if item == nil {
			continue
		}
		d.StreamListItem(ctx, instancePrivateNICRow{*item, parent.Project, parent.Organization})

		if d.RowsRemaining(ctx) == 0 {
			return nil, nil
		}
	}

	return nil, nil
}
