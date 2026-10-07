package scaleway

import (
	"net/http"
	"testing"

	awssdkerr "github.com/aws/aws-sdk-go/aws/awserr"
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

func TestIsHTTPCodeErrorAWSRequestFailure(t *testing.T) {
	reqErr := awssdkerr.NewRequestFailure(
		awssdkerr.New("AccessDenied", "Access Denied", nil),
		http.StatusForbidden,
		"request-id",
	)

	if !isHTTPCodeError(reqErr, http.StatusForbidden) {
		t.Errorf("isHTTPCodeError() = false, want true for awserr.RequestFailure with matching status code")
	}

	if isHTTPCodeError(reqErr, http.StatusNotFound) {
		t.Errorf("isHTTPCodeError() = true, want false for awserr.RequestFailure with non-matching status code")
	}
}

type errPlain string

func (e errPlain) Error() string { return string(e) }
