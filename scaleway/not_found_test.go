package scaleway

import (
	"net/http"
	"testing"

	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/smithy-go"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func TestIs403Error(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "scw.ResponseError with matching status code",
			err:  &scw.ResponseError{StatusCode: http.StatusForbidden},
			want: true,
		},
		{
			name: "scw.PermissionsDeniedError with no StatusCode field",
			err:  &scw.PermissionsDeniedError{},
			want: true,
		},
		{
			name: "scw.ResponseError with non-matching status code",
			err:  &scw.ResponseError{StatusCode: http.StatusNotFound},
			want: false,
		},
		{
			name: "unrelated error",
			err:  errPlain("boom"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := is403Error(tt.err); got != tt.want {
				t.Errorf("is403Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIs404Error(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "scw.ResponseError with matching status code",
			err:  &scw.ResponseError{StatusCode: http.StatusNotFound},
			want: true,
		},
		{
			name: "scw.ResourceNotFoundError regression check",
			err:  &scw.ResourceNotFoundError{},
			want: true,
		},
		{
			name: "scw.ResponseError with non-matching status code",
			err:  &scw.ResponseError{StatusCode: http.StatusForbidden},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := is404Error(tt.err); got != tt.want {
				t.Errorf("is404Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsNotFoundOrUnsupportedZoneError(t *testing.T) {
	if !isNotFoundOrUnsupportedZoneError(&scw.ResponseError{StatusCode: http.StatusNotFound}) {
		t.Errorf("isNotFoundOrUnsupportedZoneError() = false, want true for a 404")
	}

	if !isNotFoundOrUnsupportedZoneError(&scw.ResponseError{StatusCode: http.StatusNotImplemented}) {
		t.Errorf("isNotFoundOrUnsupportedZoneError() = false, want true for a 501")
	}

	for _, err := range []error{
		&scw.ResponseError{StatusCode: http.StatusInternalServerError},
		&scw.ResponseError{StatusCode: http.StatusForbidden},
		nil,
	} {
		if isNotFoundOrUnsupportedZoneError(err) {
			t.Errorf("isNotFoundOrUnsupportedZoneError(%v) = true, want false", err)
		}
	}
}

func TestIsHTTPCodeErrorAWSResponseError(t *testing.T) {
	respErr := &smithy.OperationError{
		ServiceID:     "S3",
		OperationName: "GetBucketAcl",
		Err: &awshttp.ResponseError{
			ResponseError: &smithyhttp.ResponseError{
				Response: &smithyhttp.Response{Response: &http.Response{StatusCode: http.StatusForbidden}},
				Err:      errPlain("AccessDenied"),
			},
			RequestID: "request-id",
		},
	}

	if !isHTTPCodeError(respErr, http.StatusForbidden) {
		t.Errorf("isHTTPCodeError() = false, want true for awshttp.ResponseError with matching status code")
	}

	if isHTTPCodeError(respErr, http.StatusNotFound) {
		t.Errorf("isHTTPCodeError() = true, want false for awshttp.ResponseError with non-matching status code")
	}
}

type errPlain string

func (e errPlain) Error() string { return string(e) }
