package cli

import (
	"fmt"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mizdebsk/radii/internal/api"
	"github.com/mizdebsk/radii/internal/mocks"
)

func TestAutoDetectForce(t *testing.T) {
	for _, force := range []bool{false, true} {
		for _, batch := range []bool{false, true} {
			for _, dryRun := range []bool{false, true} {
				t.Run(fmt.Sprintf("force=%t/batch=%t/dry=%t", force, batch, dryRun), func(t *testing.T) {
					ctrl := gomock.NewController(t)
					rm := mocks.NewMockRepositoryManager(ctrl)
					pm := mocks.NewMockPackageManager(ctrl)
					nv := mocks.NewMockProvider(ctrl)
					am := mocks.NewMockProvider(ctrl)
					absent := mocks.NewMockProvider(ctrl)
					deps := api.CoreDeps{RepositoryManager: rm, PackageManager: pm, Providers: []api.Provider{nv, am, absent}}
					var detections []*gomock.Call
					for i, p := range []*mocks.MockProvider{nv, am, absent} {
						id := []string{"nvidia", "amdgpu", "absent"}[i]
						p.EXPECT().GetID().Return(id).AnyTimes()
						p.EXPECT().GetName().Return(id).AnyTimes()
						detections = append(detections, p.EXPECT().DetectHardware().Return(i < 2, nil))
					}
					args := []string{"radii", "install", "--auto-detect"}
					if batch {
						args = append(args, "--batch")
					}
					if dryRun {
						args = append(args, "--dry-run")
					}
					if force {
						args = append(args, "--force")
						nv.EXPECT().RequiredChannels().Return([]string{"BaseOS", "Supplementary"})
						am.EXPECT().RequiredChannels().Return([]string{"BaseOS", "Extensions"})
						channels := []string{"BaseOS", "Supplementary", "Extensions"}
						var prepared, configured *gomock.Call
						if dryRun {
							repos := []string{"base", "supplementary", "extensions"}
							prepared = rm.EXPECT().GetRepoIDs(channels).Return(repos, nil)
							configured = pm.EXPECT().SetEnableRepos(repos).After(prepared)
						} else {
							prepared = rm.EXPECT().EnsureRepositoriesEnabled(channels).Return(nil)
							configured = pm.EXPECT().SetEnableRepos(gomock.Nil()).After(prepared)
						}
						for _, detected := range detections {
							prepared.After(detected)
						}
						transaction := pm.EXPECT().Install([]string{"nvidia-pkg", "amdgpu-pkg"}, batch, dryRun).Return(nil)
						for i, p := range []*mocks.MockProvider{nv, am} {
							id := []string{"nvidia", "amdgpu"}[i]
							driver := api.DriverID{ProviderID: id, Version: "default"}
							listed := p.EXPECT().ListAvailable().After(configured).Return([]api.DriverID{driver, {ProviderID: id, Version: "older"}}, nil)
							installed := p.EXPECT().Install([]api.DriverID{driver}).After(listed).Return([]string{id + "-pkg"}, nil)
							transaction.After(installed)
						}
					}
					err := Execute(args, deps, "test")
					if force {
						if err != nil {
							t.Fatal(err)
						}
					} else if err == nil || !strings.Contains(err.Error(), "radii install --auto-detect --force") {
						t.Fatalf("expected multi-provider error with force command, got %v", err)
					}
				})
			}
		}
	}
}

func TestAutoDetectForceRejectsExplicitDrivers(t *testing.T) {
	err := Execute([]string{"radii", "install", "--auto-detect", "--force", "amdgpu:latest"}, api.CoreDeps{}, "test")
	if err == nil || !strings.Contains(err.Error(), "both --auto-detect and specific drivers given") {
		t.Fatalf("expected incompatible arguments error, got %v", err)
	}
}
