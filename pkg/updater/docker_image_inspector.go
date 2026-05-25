package updater

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"
)

type DockerImageInspector struct {
	execCommand func(context.Context, string, ...string) *exec.Cmd
}

type dockerImageInspect struct {
	ID          string   `json:"Id"`
	RepoDigests []string `json:"RepoDigests"`
}

type dockerContainerInspect struct {
	Image string `json:"Image"`
}

func NewDockerImageInspector() *DockerImageInspector {
	return &DockerImageInspector{execCommand: exec.CommandContext}
}

func (i *DockerImageInspector) Inspect(ctx context.Context, image string) (ImageIdentity, error) {
	cmd := i.execCommand(ctx, "docker", "image", "inspect", image)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if isImageNotFound(stderr.String()) {
			return missingImageIdentity(image), nil
		}
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			return missingImageIdentity(image), fmt.Errorf("inspect image %s: %w", image, err)
		}
		return missingImageIdentity(image), fmt.Errorf("inspect image %s: %s", image, message)
	}

	var inspected []dockerImageInspect
	if err := json.Unmarshal(stdout.Bytes(), &inspected); err != nil {
		return missingImageIdentity(image), fmt.Errorf("parse image inspect %s: %w", image, err)
	}
	if len(inspected) == 0 {
		return missingImageIdentity(image), fmt.Errorf("inspect image %s returned no data", image)
	}

	value, source := inspectIdentityValue(inspected[0])
	if value == "" {
		return missingImageIdentity(image), fmt.Errorf("image %s has no repo digest or image ID", image)
	}

	return ImageIdentity{Image: image, Value: value, ID: inspected[0].ID, Present: true, Source: source}, nil
}

func (i *DockerImageInspector) InspectContainer(ctx context.Context, container string) (ContainerIdentity, error) {
	cmd := i.execCommand(ctx, "docker", "container", "inspect", container)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if isImageNotFound(stderr.String()) {
			return ContainerIdentity{Container: container}, nil
		}
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			return ContainerIdentity{Container: container}, fmt.Errorf("inspect container %s: %w", container, err)
		}
		return ContainerIdentity{Container: container}, fmt.Errorf("inspect container %s: %s", container, message)
	}

	var inspected []dockerContainerInspect
	if err := json.Unmarshal(stdout.Bytes(), &inspected); err != nil {
		return ContainerIdentity{Container: container}, fmt.Errorf("parse container inspect %s: %w", container, err)
	}
	if len(inspected) == 0 || inspected[0].Image == "" {
		return ContainerIdentity{Container: container}, fmt.Errorf("container %s has no image ID", container)
	}

	return ContainerIdentity{Container: container, ImageID: inspected[0].Image, Present: true}, nil
}

func inspectIdentityValue(inspected dockerImageInspect) (string, string) {
	if digest := firstRepoDigest(inspected.RepoDigests); digest != "" {
		return digest, IdentitySourceRepoDigest
	}
	if inspected.ID != "" {
		return inspected.ID, IdentitySourceImageID
	}
	return "", ""
}

func firstRepoDigest(repoDigests []string) string {
	sorted := append([]string(nil), repoDigests...)
	sort.Strings(sorted)

	for _, repoDigest := range sorted {
		_, digest, found := strings.Cut(repoDigest, "@")
		if found && digest != "" {
			return digest
		}
		if repoDigest != "" {
			return repoDigest
		}
	}

	return ""
}

func missingImageIdentity(image string) ImageIdentity {
	return ImageIdentity{Image: image, Value: IdentityNotPresent, Present: false, Source: IdentitySourceMissing}
}

func isImageNotFound(stderr string) bool {
	lower := strings.ToLower(stderr)
	return strings.Contains(lower, "no such image") ||
		strings.Contains(lower, "no such container") ||
		strings.Contains(lower, "no such object") ||
		strings.Contains(lower, "not found")
}
