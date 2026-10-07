package scaleway

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

// getSessionConfig :: returns Scaleway client to perform API requests
func getSessionConfig(ctx context.Context, d *plugin.QueryData) (*scw.Client, error) {
	// Load clientOptions from cache
	sessionCacheKey := "scaleway.clientoption"
	if cachedData, ok := d.ConnectionManager.Cache.Get(sessionCacheKey); ok {
		return cachedData.(*scw.Client), nil
	}

	var accessKey, secretKey string

	opts := []scw.ClientOption{}

	// Load credentials from environment variables
	loadEnv := scw.LoadEnvProfile()

	if loadEnv.AccessKey != nil && loadEnv.SecretKey != nil {
		accessKey = *loadEnv.AccessKey
		secretKey = *loadEnv.SecretKey
	}

	// Get scaleway config
	scalewayConfig := GetConfig(d.Connection)

	if scalewayConfig.AccessKey != nil && scalewayConfig.SecretKey == nil {
		return nil, fmt.Errorf("partial credentials found in connection config, missing: secret_key")
	} else if scalewayConfig.SecretKey != nil && scalewayConfig.AccessKey == nil {
		return nil, fmt.Errorf("partial credentials found in connection config, missing: access_key")
	} else if scalewayConfig.AccessKey != nil && scalewayConfig.SecretKey != nil {
		accessKey = *scalewayConfig.AccessKey
		secretKey = *scalewayConfig.SecretKey
	}

	// No creds
	if accessKey == "" && secretKey == "" {
		return nil, fmt.Errorf("both access_key and secret_key must be configured")
	}

	opts = append(opts, scw.WithAuth(accessKey, secretKey))

	// Create client
	client, err := scw.NewClient(opts...)
	if err != nil {
		return nil, err
	}

	// save clientOptions in cache
	d.ConnectionManager.Cache.Set(sessionCacheKey, client)

	return client, nil
}

// objectAccessKey returns the S3 access key used to scope Object Storage requests to a project.
// Scaleway's S3 API is project-scoped; the project is selected by suffixing the access key with `@<project_id>`.
func objectAccessKey(accessKey, project string) string {
	if project == "" {
		return accessKey
	}

	return accessKey + "@" + project
}

// objectSessionCacheKey returns the cache key of the S3 client for a region and an optional project.
func objectSessionCacheKey(region, project string) string {
	if project == "" {
		return "scaleway.objectclient-" + region
	}

	return "scaleway.objectclient-" + region + "-" + project
}

// getObjectSessionConfig :: returns S3 client to perform Object Storage API requests.
// An empty project uses the default project of the credentials.
func getObjectSessionConfig(ctx context.Context, d *plugin.QueryData, region, project string) (*s3.S3, error) {
	// Load clientOptions from cache
	sessionCacheKey := objectSessionCacheKey(region, project)
	if cachedData, ok := d.ConnectionManager.Cache.Get(sessionCacheKey); ok {
		return cachedData.(*s3.S3), nil
	}

	var accessKey, secretKey string

	// Load credentials from environment variables
	loadEnv := scw.LoadEnvProfile()

	if loadEnv.AccessKey != nil && loadEnv.SecretKey != nil {
		accessKey = *loadEnv.AccessKey
		secretKey = *loadEnv.SecretKey
	}

	// Get scaleway config
	scalewayConfig := GetConfig(d.Connection)

	if scalewayConfig.AccessKey != nil && scalewayConfig.SecretKey == nil {
		return nil, fmt.Errorf("partial credentials found in connection config, missing: secret_key")
	} else if scalewayConfig.SecretKey != nil && scalewayConfig.AccessKey == nil {
		return nil, fmt.Errorf("partial credentials found in connection config, missing: access_key")
	} else if scalewayConfig.AccessKey != nil && scalewayConfig.SecretKey != nil {
		accessKey = *scalewayConfig.AccessKey
		secretKey = *scalewayConfig.SecretKey
	}

	// No creds
	if accessKey == "" && secretKey == "" {
		return nil, fmt.Errorf("both access_key and secret_key must be configured")
	}

	// session default configuration
	sessionOptions := session.Options{
		Config: aws.Config{
			Region:      &region,
			Credentials: credentials.NewStaticCredentials(objectAccessKey(accessKey, project), secretKey, ""),
			Endpoint:    scw.StringPtr("https://s3." + region + ".scw.cloud"),
		},
	}

	s, err := session.NewSessionWithOptions(sessionOptions)
	if err != nil {
		return nil, err
	}
	client := s3.New(s)

	// save clientOptions in cache
	d.ConnectionManager.Cache.Set(sessionCacheKey, client)

	return client, nil
}
