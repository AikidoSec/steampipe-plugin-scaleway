package scaleway

import "testing"

func TestObjectAccessKey(t *testing.T) {
	tests := []struct {
		name    string
		project string
		want    string
	}{
		{name: "no project uses the bare access key", project: "", want: "SCWXXXX"},
		{name: "project is appended to the access key", project: "proj-1", want: "SCWXXXX@proj-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := objectAccessKey("SCWXXXX", tt.project); got != tt.want {
				t.Errorf("objectAccessKey() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestObjectSessionCacheKey(t *testing.T) {
	keys := map[string]struct{}{
		objectSessionCacheKey("fr-par", ""):       {},
		objectSessionCacheKey("nl-ams", ""):       {},
		objectSessionCacheKey("fr-par", "proj-1"): {},
		objectSessionCacheKey("fr-par", "proj-2"): {},
	}
	if len(keys) != 4 {
		t.Errorf("expected distinct cache keys per region and project, got %d distinct keys", len(keys))
	}

	if objectSessionCacheKey("fr-par", "proj-1") != objectSessionCacheKey("fr-par", "proj-1") {
		t.Error("cache key must be stable for equal inputs")
	}
}

func TestBucketProject(t *testing.T) {
	owner := func(s string) *string { return &s }

	tests := []struct {
		name          string
		scopedProject string
		ownerID       *string
		want          string
	}{
		{name: "scoped project wins over owner", scopedProject: "proj-1", ownerID: owner("org:proj-2"), want: "proj-1"},
		{name: "unscoped uses the project part of the owner", ownerID: owner("org:proj-2"), want: "proj-2"},
		{name: "owner without separator does not panic", ownerID: owner("org"), want: ""},
		{name: "nil owner does not panic", ownerID: nil, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := bucketProject(tt.scopedProject, tt.ownerID); got != tt.want {
				t.Errorf("bucketProject() = %q, want %q", got, tt.want)
			}
		})
	}
}
