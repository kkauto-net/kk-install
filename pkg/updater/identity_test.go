package updater

import (
	"context"
	"errors"
	"testing"
)

type fakeInspector map[string]ImageIdentity

func (f fakeInspector) Inspect(_ context.Context, image string) (ImageIdentity, error) {
	identity, ok := f[image]
	if !ok {
		return ImageIdentity{}, errors.New("unexpected image: " + image)
	}
	return identity, nil
}

type fakeContainerInspector map[string]ContainerIdentity

func (f fakeContainerInspector) InspectContainer(_ context.Context, container string) (ContainerIdentity, error) {
	identity, ok := f[container]
	if !ok {
		return ContainerIdentity{}, errors.New("unexpected container: " + container)
	}
	return identity, nil
}

func TestSnapshotImages(t *testing.T) {
	inspector := fakeInspector{
		"redis:7-alpine": {Value: "sha256:abc", Present: true, Source: IdentitySourceRepoDigest},
	}

	snapshot, err := SnapshotImages(context.Background(), []string{"redis:7-alpine"}, inspector)
	if err != nil {
		t.Fatalf("SnapshotImages() error = %v", err)
	}

	identity := snapshot["redis:7-alpine"]
	if identity.Image != "redis:7-alpine" {
		t.Fatalf("Image = %q, want redis:7-alpine", identity.Image)
	}
	if identity.Value != "sha256:abc" {
		t.Fatalf("Value = %q, want sha256:abc", identity.Value)
	}
}

func TestCompareSnapshots(t *testing.T) {
	tests := []struct {
		name        string
		before      map[string]ImageIdentity
		after       map[string]ImageIdentity
		wantUpdates int
		wantOld     string
		wantNew     string
		wantErr     bool
	}{
		{
			name: "unchanged digest",
			before: map[string]ImageIdentity{
				"redis:7-alpine": {Image: "redis:7-alpine", Value: "sha256:abc", Present: true},
			},
			after: map[string]ImageIdentity{
				"redis:7-alpine": {Image: "redis:7-alpine", Value: "sha256:abc", Present: true},
			},
		},
		{
			name: "changed digest",
			before: map[string]ImageIdentity{
				"redis:7-alpine": {Image: "redis:7-alpine", Value: "sha256:abc", Present: true},
			},
			after: map[string]ImageIdentity{
				"redis:7-alpine": {Image: "redis:7-alpine", Value: "sha256:def", Present: true},
			},
			wantUpdates: 1,
			wantOld:     "sha256:abc",
			wantNew:     "sha256:def",
		},
		{
			name: "missing before present after",
			before: map[string]ImageIdentity{
				"redis:7-alpine": {Image: "redis:7-alpine", Value: IdentityNotPresent, Present: false},
			},
			after: map[string]ImageIdentity{
				"redis:7-alpine": {Image: "redis:7-alpine", Value: "sha256:def", Present: true},
			},
			wantUpdates: 1,
			wantOld:     IdentityNotPresent,
			wantNew:     "sha256:def",
		},
		{
			name: "missing after errors",
			before: map[string]ImageIdentity{
				"redis:7-alpine": {Image: "redis:7-alpine", Value: "sha256:abc", Present: true},
			},
			after: map[string]ImageIdentity{
				"redis:7-alpine": {Image: "redis:7-alpine", Value: IdentityNotPresent, Present: false},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updates, err := CompareSnapshots(tt.before, tt.after)
			if tt.wantErr {
				if err == nil {
					t.Fatal("CompareSnapshots() expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("CompareSnapshots() error = %v", err)
			}
			if len(updates) != tt.wantUpdates {
				t.Fatalf("len(updates) = %d, want %d", len(updates), tt.wantUpdates)
			}
			if tt.wantUpdates == 0 {
				return
			}
			if updates[0].OldDigest != tt.wantOld || updates[0].NewDigest != tt.wantNew {
				t.Fatalf("update = %#v, want old %q new %q", updates[0], tt.wantOld, tt.wantNew)
			}
		})
	}
}

func TestCompareRunningContainers(t *testing.T) {
	targets := []ContainerTarget{{Image: "redis:7-alpine", Container: "kkengine_redis"}}
	local := map[string]ImageIdentity{
		"redis:7-alpine": {Image: "redis:7-alpine", Value: "sha256:newdigest", ID: "sha256:newid", Present: true},
	}

	t.Run("running old image needs recreate", func(t *testing.T) {
		updates, err := CompareRunningContainers(context.Background(), targets, local, fakeContainerInspector{
			"kkengine_redis": {Container: "kkengine_redis", ImageID: "sha256:oldid", Present: true},
		})
		if err != nil {
			t.Fatalf("CompareRunningContainers() error = %v", err)
		}
		if len(updates) != 1 {
			t.Fatalf("len(updates) = %d, want 1", len(updates))
		}
		if updates[0].OldDigest != "sha256:oldid" || updates[0].NewDigest != "sha256:newdigest" {
			t.Fatalf("update = %#v", updates[0])
		}
	})

	t.Run("matching running image is current", func(t *testing.T) {
		updates, err := CompareRunningContainers(context.Background(), targets, local, fakeContainerInspector{
			"kkengine_redis": {Container: "kkengine_redis", ImageID: "sha256:newid", Present: true},
		})
		if err != nil {
			t.Fatalf("CompareRunningContainers() error = %v", err)
		}
		if len(updates) != 0 {
			t.Fatalf("len(updates) = %d, want 0", len(updates))
		}
	})

	t.Run("missing container is ignored", func(t *testing.T) {
		updates, err := CompareRunningContainers(context.Background(), targets, local, fakeContainerInspector{
			"kkengine_redis": {Container: "kkengine_redis", Present: false},
		})
		if err != nil {
			t.Fatalf("CompareRunningContainers() error = %v", err)
		}
		if len(updates) != 0 {
			t.Fatalf("len(updates) = %d, want 0", len(updates))
		}
	})
}

func TestMergeImageUpdates(t *testing.T) {
	primary := []ImageUpdate{{Image: "redis:7-alpine", OldDigest: "old", NewDigest: "new"}}
	secondary := []ImageUpdate{
		{Image: "redis:7-alpine", OldDigest: "container-old", NewDigest: "new"},
		{Image: "mariadb:10.6", OldDigest: "old-db", NewDigest: "new-db"},
	}

	merged := MergeImageUpdates(primary, secondary)
	if len(merged) != 2 {
		t.Fatalf("len(merged) = %d, want 2", len(merged))
	}
	if merged[0].Image != "redis:7-alpine" || merged[1].Image != "mariadb:10.6" {
		t.Fatalf("merged = %#v", merged)
	}
}

func TestInspectIdentityValue(t *testing.T) {
	value, source := inspectIdentityValue(dockerImageInspect{
		ID: "sha256:imageid",
		RepoDigests: []string{
			"example/app@sha256:bbb",
			"example/app@sha256:aaa",
		},
	})
	if value != "sha256:aaa" || source != IdentitySourceRepoDigest {
		t.Fatalf("inspectIdentityValue() = %q, %q", value, source)
	}

	value, source = inspectIdentityValue(dockerImageInspect{ID: "sha256:imageid"})
	if value != "sha256:imageid" || source != IdentitySourceImageID {
		t.Fatalf("inspectIdentityValue() fallback = %q, %q", value, source)
	}
}
