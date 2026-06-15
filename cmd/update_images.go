package cmd

import (
	"context"
	"fmt"

	"github.com/kkauto-net/kk-install/pkg/compose"
	"github.com/kkauto-net/kk-install/pkg/monitor"
	"github.com/kkauto-net/kk-install/pkg/ui"
	"github.com/kkauto-net/kk-install/pkg/updater"
)

type updateImageState struct {
	composeFile *compose.ComposeFile
	images      []string
	inspector   updater.ImageInspector
	containers  updater.ContainerInspector
	before      map[string]updater.ImageIdentity
}

func prepareUpdateImageState(ctx context.Context, cwd string) (*updateImageState, error) {
	composeFile, err := compose.ParseComposeFile(cwd)
	if err != nil {
		return nil, err
	}

	images := composeFile.GetServiceImages()
	if len(images) == 0 {
		return nil, fmt.Errorf("no service images defined in docker-compose.yml")
	}

	inspector := updater.NewDockerImageInspector()
	inspectCtx, cancel := context.WithTimeout(ctx, compose.DefaultTimeout)
	defer cancel()

	before, err := updater.SnapshotImages(inspectCtx, images, inspector)
	if err != nil {
		return nil, err
	}

	return &updateImageState{
		composeFile: composeFile,
		images:      images,
		inspector:   inspector,
		containers:  inspector,
		before:      before,
	}, nil
}

func detectUpdatesAfterPull(ctx context.Context, state *updateImageState) ([]updater.ImageUpdate, error) {
	inspectCtx, cancel := context.WithTimeout(ctx, compose.DefaultTimeout)
	defer cancel()

	after, err := updater.SnapshotImages(inspectCtx, state.images, state.inspector)
	if err != nil {
		return nil, err
	}

	imageUpdates, err := updater.CompareSnapshots(state.before, after)
	if err != nil {
		return nil, err
	}

	containerUpdates, err := updater.CompareRunningContainers(inspectCtx, serviceContainerTargets(state.composeFile), after, state.containers)
	if err != nil {
		return nil, err
	}

	return updater.MergeImageUpdates(imageUpdates, containerUpdates), nil
}

func serviceContainerTargets(composeFile *compose.ComposeFile) []updater.ContainerTarget {
	targets := make([]updater.ContainerTarget, 0, len(composeFile.Services))
	for _, name := range composeFile.GetServiceNames() {
		image := composeFile.Services[name].Image
		if image == "" {
			continue
		}
		targets = append(targets, updater.ContainerTarget{
			Image:     image,
			Container: composeFile.GetServiceContainerName(name),
		})
	}
	return targets
}

func showUpdatePreparationError(err error) {
	ui.ShowBoxedError(ui.ErrorSuggestion{
		Title:      ui.Msg("pull_failed"),
		Message:    ui.SanitizeError(err),
		Suggestion: ui.Msg("err_update_prepare_suggestion"),
		Command:    "kk init",
	})
}

func monitorUpdateHealth(ctx context.Context, composeFile *compose.ComposeFile) {
	healthMonitor, err := monitor.NewHealthMonitor()
	if err != nil {
		return
	}
	defer healthMonitor.Close()

	var containers []monitor.ContainerInfo
	for _, name := range composeFile.GetServiceNames() {
		containers = append(containers, monitor.ContainerInfo{
			ServiceName:    name,
			ContainerName:  composeFile.GetServiceContainerName(name),
			HasHealthCheck: composeFile.HasHealthCheck(name),
		})
	}

	healthMonitor.MonitorAll(ctx, containers, func(status monitor.HealthStatus) {
		ui.ShowServiceProgress(status.ServiceName, status.Status)
	})
}
