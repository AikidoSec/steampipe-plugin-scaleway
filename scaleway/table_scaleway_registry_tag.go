package scaleway

import (
	"context"

	"github.com/scaleway/scaleway-sdk-go/api/registry/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

//// TABLE DEFINITION

func tableScalewayRegistryTag(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:              "scaleway_registry_tag",
		Description:       "A registry tag is a named, versioned reference to a container image.",
		GetMatrixItemFunc: BuildRegionList,
		List: &plugin.ListConfig{
			Hydrate:       listRegistryTags,
			ParentHydrate: listRegistryNamespaces,
			KeyColumns: []*plugin.KeyColumn{
				{
					Name:    "region",
					Require: plugin.Optional,
				},
				{
					Name:    "namespace_id",
					Require: plugin.Optional,
				},
				{
					Name:    "image_id",
					Require: plugin.Optional,
				},
			},
		},
		Columns: []*plugin.Column{
			{
				Name:        "id",
				Description: "The unique identifier of the tag.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID"),
			},
			{
				Name:        "name",
				Description: "The name of the tag.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "image_id",
				Description: "The ID of the image the tag belongs to.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ImageID"),
			},
			{
				Name:        "namespace_id",
				Description: "The ID of the namespace the image belongs to.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("NamespaceID"),
			},
			{
				Name:        "status",
				Description: "The status of the tag.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Status").Transform(transform.ToString),
			},
			{
				Name:        "digest",
				Description: "The digest of the image manifest the tag points to.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "created_at",
				Description: "The time when the tag was created.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "updated_at",
				Description: "The time when the tag was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
			},

			// Scaleway standard columns
			{
				Name:        "region",
				Description: "The region where the tag resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Region").Transform(transform.ToString),
			},
			{
				Name:        "project_id",
				Description: "The ID of the project where the tag resides.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ProjectID"),
			},
			{
				Name:        "organization_id",
				Description: "The ID of the organization where the tag resides.",
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

type registryTagRow struct {
	registry.Tag
	NamespaceID    string
	Region         scw.Region
	ProjectID      string
	OrganizationID string
}

//// LIST FUNCTION

func listRegistryTags(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	matrixRegion := d.EqualsQualString("region")
	namespace := h.Item.(*registry.Namespace)

	if namespace.Region.String() != matrixRegion {
		return nil, nil
	}
	if id := d.EqualsQualString("namespace_id"); id != "" && id != namespace.ID {
		return nil, nil
	}

	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway_registry_tag.listRegistryTags", "connection_error", err)
		return nil, err
	}

	registryApi := registry.NewAPI(client)

	imageReq := &registry.ListImagesRequest{
		Region:      namespace.Region,
		NamespaceID: &namespace.ID,
		Page:        scw.Int32Ptr(1),
		PageSize:    scw.Uint32Ptr(maxListPageSize),
	}

	var imageCount int

	for {
		imageResp, err := registryApi.ListImages(imageReq)
		if err != nil {
			plugin.Logger(ctx).Error("scaleway_registry_tag.listRegistryTags", "image_query_error", err)
			if isNotFoundOrUnsupportedZoneError(err) {
				return nil, nil
			}
			return nil, err
		}

		for _, image := range imageResp.Images {
			imageCount++

			if id := d.EqualsQualString("image_id"); id != "" && id != image.ID {
				continue
			}

			tagReq := &registry.ListTagsRequest{
				Region:   namespace.Region,
				ImageID:  image.ID,
				Page:     scw.Int32Ptr(1),
				PageSize: scw.Uint32Ptr(listPageSize(d)),
			}

			var tagCount int

			for {
				tagResp, err := registryApi.ListTags(tagReq)
				if err != nil {
					plugin.Logger(ctx).Error("scaleway_registry_tag.listRegistryTags", "query_error", err)
					if isNotFoundOrUnsupportedZoneError(err) {
						break
					}
					return nil, err
				}

				for _, tag := range tagResp.Tags {
					d.StreamListItem(ctx, registryTagRow{*tag, namespace.ID, namespace.Region, namespace.ProjectID, namespace.OrganizationID})
					tagCount++

					if d.RowsRemaining(ctx) == 0 {
						return nil, nil
					}
				}

				if len(tagResp.Tags) == 0 || tagResp.TotalCount == uint32(tagCount) {
					break
				}
				tagReq.Page = scw.Int32Ptr(*tagReq.Page + 1)
			}
		}

		if len(imageResp.Images) == 0 || imageResp.TotalCount == uint32(imageCount) {
			break
		}
		imageReq.Page = scw.Int32Ptr(*imageReq.Page + 1)
	}

	return nil, nil
}
