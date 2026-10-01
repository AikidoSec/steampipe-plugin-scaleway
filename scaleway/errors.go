package scaleway

import (
	"context"
	"net/http"
	"strconv"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

// shouldIgnoreErrors returns an ErrorPredicate that matches when the error's
// HTTP status code is one of statusCodes, or matches a status code
// configured via the connection's "ignore_error_codes".
func shouldIgnoreErrors(statusCodes []int) plugin.ErrorPredicateWithContext {
	return func(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData, err error) bool {
		config := GetConfig(d.Connection)
		allCodes := append(append([]int{}, statusCodes...), parseStatusCodes(config.IgnoreErrorCodes)...)
		return matchesAnyStatusCode(err, allCodes)
	}
}

// shouldIgnoreErrorPluginDefault returns an ErrorPredicate based solely on
// the connection's "ignore_error_codes", for use as the plugin's
// DefaultIgnoreConfig.
func shouldIgnoreErrorPluginDefault() plugin.ErrorPredicateWithContext {
	return func(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData, err error) bool {
		config := GetConfig(d.Connection)
		if len(config.IgnoreErrorCodes) == 0 {
			return false
		}
		return matchesAnyStatusCode(err, parseStatusCodes(config.IgnoreErrorCodes))
	}
}

// matchesAnyStatusCode checks err against each status code, using the
// typed-error-aware predicates for 403/404 and the raw status-code check
// otherwise.
func matchesAnyStatusCode(err error, statusCodes []int) bool {
	for _, code := range statusCodes {
		switch code {
		case http.StatusForbidden:
			if is403Error(err) {
				return true
			}
		case http.StatusNotFound:
			if is404Error(err) {
				return true
			}
		default:
			if isHTTPCodeError(err, code) {
				return true
			}
		}
	}
	return false
}

// parseStatusCodes converts the connection config's plain status-code
// strings (e.g. "403") to ints, silently skipping any that don't parse.
func parseStatusCodes(codes []string) []int {
	var statusCodes []int
	for _, code := range codes {
		if parsed, err := strconv.Atoi(code); err == nil {
			statusCodes = append(statusCodes, parsed)
		}
	}
	return statusCodes
}
