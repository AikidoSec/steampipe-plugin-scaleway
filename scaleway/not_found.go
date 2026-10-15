package scaleway

import (
	"errors"
	"net/http"

	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func is404Error(err error) bool {
	notFoundError := &scw.ResourceNotFoundError{}
	return isHTTPCodeError(err, http.StatusNotFound) || errors.As(err, &notFoundError)
}

func is403Error(err error) bool {
	permissionsDeniedError := &scw.PermissionsDeniedError{}
	return isHTTPCodeError(err, http.StatusForbidden) || errors.As(err, &permissionsDeniedError)
}

// isHTTPCodeError matches statusCode against either a native-API
// *scw.ResponseError or an S3-compatible-endpoint *awshttp.ResponseError.
func isHTTPCodeError(err error, statusCode int) bool {
	if err == nil {
		return false
	}

	responseError := &scw.ResponseError{}
	if errors.As(err, &responseError) && responseError.StatusCode == statusCode {
		return true
	}

	var awsResponseError *awshttp.ResponseError
	if errors.As(err, &awsResponseError) && awsResponseError.HTTPStatusCode() == statusCode {
		return true
	}

	return false
}
