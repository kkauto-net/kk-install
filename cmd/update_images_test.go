package cmd

import (
	"context"
	"errors"
	"testing"

	"github.com/kkauto-net/kk-install/pkg/compose"
	"github.com/kkauto-net/kk-install/pkg/updater"
)

type fakeUpdateImageInspector map[string]updater.ImageIdentity

func (f fakeUpdateImageInspector) Inspect(_ context.Context, image string) (updater.ImageIdentity, error) {
	identity, ok := f[image]
	if !ok {
		return updater.ImageIdentity{}, errors.New("unexpected image: " + image)
	}
	return identity, nil
}

type fakeUpdateContainerInspector map[string]updater.ContainerIdentity

func (f fakeUpdateContainerInspector) InspectContainer(_ context.Context, container string) (updater.ContainerIdentity, error) {
	identity, ok := f[container]
	if !ok {
		return updater.ContainerIdentity{}, errors.New("unexpected container: " + container)
	}
	return identity, nil
}

func TestDetectUpdatesAfterPullIncludesStaleRunningContainer(t *testing.T) {
	state := &updateImageState{
		composeFile: &compose.ComposeFile{
			Services: map[string]compose.Service{
				"redis": {Image: "redis:alpine", ContainerName: "kkengine_redis"},
			},
		},
		images: []string{"redis:alpine"},
		before: map[string]updater.ImageIdentity{
			"redis:alpine": {Image: "redis:alpine", Value: "sha256:newdigest", ID: "sha256:newid", Present: true},
		},
		inspector: fakeUpdateImageInspector{
			"redis:alpine": {Image: "redis:alpine", Value: "sha256:newdigest", ID: "sha256:newid", Present: true},
		},
		containers: fakeUpdateContainerInspector{
			"kkengine_redis": {Container: "kkengine_redis", ImageID: "sha256:oldid", Present: true},
		},
	}

	updates, err := detectUpdatesAfterPull(context.Background(), state)
	if err != nil {
		t.Fatalf("detectUpdatesAfterPull() error = %v", err)
	}
	if len(updates) != 1 {
		t.Fatalf("len(updates) = %d, want 1", len(updates))
	}
	if updates[0].Image != "redis:alpine" || updates[0].OldDigest != "sha256:oldid" || updates[0].NewDigest != "sha256:newdigest" {
		t.Fatalf("update = %#v", updates[0])
	}
}
