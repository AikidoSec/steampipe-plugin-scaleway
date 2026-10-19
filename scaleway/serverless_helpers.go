package scaleway

import (
	"context"
	"net/url"
	"strings"

	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

// cachedParentProjectID returns the project ID resolved by fetch for a row whose SDK struct has no project field.
// The result is cached per cacheKey; a 403/404 from fetch is cached and yields nil.
func cachedParentProjectID(ctx context.Context, d *plugin.QueryData, cacheKey string, fetch func(client *scw.Client) (string, error)) (interface{}, error) {
	if cached, ok := d.ConnectionManager.Cache.Get(cacheKey); ok {
		projectID, _ := cached.(string)
		if projectID == "" {
			return nil, nil
		}
		return projectID, nil
	}

	client, err := getSessionConfig(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("scaleway.cachedParentProjectID", "connection_error", err)
		return nil, err
	}

	projectID, err := fetch(client)
	if err != nil {
		if is403Error(err) || isNotFoundOrUnsupportedZoneError(err) {
			d.ConnectionManager.Cache.Set(cacheKey, "")
			return nil, nil
		}
		plugin.Logger(ctx).Error("scaleway.cachedParentProjectID", "query_error", err)
		return nil, err
	}

	d.ConnectionManager.Cache.Set(cacheKey, projectID)
	if projectID == "" {
		return nil, nil
	}
	return projectID, nil
}

// redactURLUserInfo removes any user info from a URL or scheme-less host[:port] value;
// a value that cannot be parsed is returned as an empty string.
func redactURLUserInfo(raw string) string {
	if raw == "" {
		return ""
	}
	schemeless := !strings.Contains(raw, "://")
	toParse := raw
	if schemeless {
		toParse = "//" + raw
	}
	u, err := url.Parse(toParse)
	if err != nil {
		return ""
	}
	u.User = nil
	out := u.String()
	if schemeless {
		out = strings.TrimPrefix(out, "//")
	}
	return out
}

// redactURLUserInfoTransform applies redactURLUserInfo to a string column value.
func redactURLUserInfoTransform(_ context.Context, d *transform.TransformData) (interface{}, error) {
	raw, ok := d.Value.(string)
	if !ok || raw == "" {
		return nil, nil
	}
	return redactURLUserInfo(raw), nil
}
