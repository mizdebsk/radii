package core

import (
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mizdebsk/radii/internal/api"
	"github.com/mizdebsk/radii/internal/mocks"
)

func TestDryRunUsesQueryRepositories(t *testing.T) {
	for _, auto := range []bool{false, true} {
		for _, batch := range []bool{false, true} {
			for _, missing := range []bool{false, true} {
				t.Run(fmt.Sprintf("auto=%t/batch=%t/missing=%t", auto, batch, missing), func(t *testing.T) {
					ctrl := gomock.NewController(t)
					rm := mocks.NewMockRepositoryManager(ctrl)
					pm := mocks.NewMockPackageManager(ctrl)
					p := mocks.NewMockProvider(ctrl)
					p.EXPECT().GetID().Return("amdgpu").AnyTimes()
					p.EXPECT().GetName().Return("AMD GPU").AnyTimes()
					if auto {
						p.EXPECT().DetectHardware().Return(true, nil)
					}
					p.EXPECT().RequiredChannels().Return([]string{"BaseOS", "Extensions"})
					query := rm.EXPECT().GetRepoIDs([]string{"BaseOS", "Extensions"})
					missingErr := errors.New("missing repository definitions")
					if missing {
						query.Return(nil, missingErr)
					} else {
						query.Return([]string{"base-repo", "extensions-repo"}, nil)
						configured := pm.EXPECT().SetEnableRepos([]string{"base-repo", "extensions-repo"}).After(query)
						drivers := []api.DriverID{{ProviderID: "amdgpu", Version: "latest"}}
						listed := p.EXPECT().ListAvailable().After(configured).Return(drivers, nil)
						installed := p.EXPECT().Install(drivers).After(listed).Return([]string{"kmod-amdgpu", "rocm-devel"}, nil)
						pm.EXPECT().Install([]string{"kmod-amdgpu", "rocm-devel"}, batch, true).After(installed).Return(nil)
					}
					deps := api.CoreDeps{Providers: []api.Provider{p}, RepositoryManager: rm, PackageManager: pm}
					var err error
					if auto {
						err = InstallAutoDetect(deps, batch, true, false)
					} else {
						err = InstallSpecific(deps, []string{"amdgpu:latest"}, batch, true, true)
					}
					if missing {
						if !errors.Is(err, missingErr) {
							t.Fatalf("expected underlying repository error, got %v", err)
						}
					} else if err != nil {
						t.Fatal(err)
					}
				})
			}
		}
	}
}
