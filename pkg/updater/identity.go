package updater

import (
	"context"
	"fmt"
	"sort"
)

const (
	IdentityNotPresent       = "-"
	IdentitySourceRepoDigest = "repo-digest"
	IdentitySourceImageID    = "image-id"
	IdentitySourceMissing    = "missing"
)

type ImageIdentity struct {
	Image   string
	Value   string
	ID      string
	Present bool
	Source  string
}

type ImageInspector interface {
	Inspect(ctx context.Context, image string) (ImageIdentity, error)
}

type ContainerIdentity struct {
	Container string
	ImageID   string
	Present   bool
}

type ContainerTarget struct {
	Image     string
	Container string
}

type ContainerInspector interface {
	InspectContainer(ctx context.Context, container string) (ContainerIdentity, error)
}

func SnapshotImages(ctx context.Context, images []string, inspector ImageInspector) (map[string]ImageIdentity, error) {
	if inspector == nil {
		return nil, fmt.Errorf("image inspector is nil")
	}

	snapshot := make(map[string]ImageIdentity, len(images))
	for _, image := range images {
		identity, err := inspector.Inspect(ctx, image)
		if err != nil {
			return nil, err
		}
		if identity.Image == "" {
			identity.Image = image
		}
		snapshot[image] = identity
	}

	return snapshot, nil
}

func CompareSnapshots(before, after map[string]ImageIdentity) ([]ImageUpdate, error) {
	images := sortedSnapshotImages(before, after)
	updates := make([]ImageUpdate, 0)

	for _, image := range images {
		afterIdentity, ok := after[image]
		if !ok || !afterIdentity.Present {
			return nil, fmt.Errorf("image %s is not present after pull", image)
		}
		if afterIdentity.Value == "" {
			return nil, fmt.Errorf("image %s has empty identity after pull", image)
		}

		beforeIdentity := before[image]
		oldValue := beforeIdentity.Value
		if !beforeIdentity.Present || oldValue == "" {
			oldValue = IdentityNotPresent
		}

		if !beforeIdentity.Present || oldValue != afterIdentity.Value {
			updates = append(updates, ImageUpdate{
				Image:     image,
				OldDigest: oldValue,
				NewDigest: afterIdentity.Value,
				Updated:   true,
			})
		}
	}

	return updates, nil
}

func CompareRunningContainers(ctx context.Context, targets []ContainerTarget, local map[string]ImageIdentity, inspector ContainerInspector) ([]ImageUpdate, error) {
	if inspector == nil {
		return nil, fmt.Errorf("container inspector is nil")
	}

	updates := make([]ImageUpdate, 0)
	seen := make(map[string]bool, len(targets))

	for _, target := range targets {
		if target.Image == "" || target.Container == "" || seen[target.Image] {
			continue
		}

		desired, ok := local[target.Image]
		if !ok || !desired.Present {
			return nil, fmt.Errorf("image %s is not present after pull", target.Image)
		}
		if desired.ID == "" {
			return nil, fmt.Errorf("image %s has empty image ID after pull", target.Image)
		}

		running, err := inspector.InspectContainer(ctx, target.Container)
		if err != nil {
			return nil, err
		}
		if !running.Present {
			continue
		}
		if running.ImageID != desired.ID {
			updates = append(updates, ImageUpdate{
				Image:     target.Image,
				OldDigest: running.ImageID,
				NewDigest: desired.Value,
				Updated:   true,
			})
			seen[target.Image] = true
		}
	}

	return updates, nil
}

func sortedSnapshotImages(before, after map[string]ImageIdentity) []string {
	seen := make(map[string]bool, len(before)+len(after))
	images := make([]string, 0, len(before)+len(after))

	for image := range before {
		seen[image] = true
		images = append(images, image)
	}
	for image := range after {
		if !seen[image] {
			images = append(images, image)
		}
	}
	sort.Strings(images)

	return images
}

func MergeImageUpdates(primary, secondary []ImageUpdate) []ImageUpdate {
	merged := append([]ImageUpdate(nil), primary...)
	seen := make(map[string]bool, len(primary)+len(secondary))
	for _, update := range primary {
		seen[update.Image] = true
	}
	for _, update := range secondary {
		if seen[update.Image] {
			continue
		}
		merged = append(merged, update)
		seen[update.Image] = true
	}
	return merged
}
