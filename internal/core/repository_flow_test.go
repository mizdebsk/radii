package core

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mizdebsk/radii/internal/api"
	"github.com/mizdebsk/radii/internal/mocks"
)

func TestInstallValidatesAllArgumentsBeforeRepositorySetup(t *testing.T) {
	for _, invalid := range []string{"invalid-format", "unknown:1"} {
		t.Run(invalid, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			provider := mocks.NewMockProvider(ctrl)
			provider.EXPECT().GetID().Return("amdgpu").AnyTimes()
			deps := api.CoreDeps{
				Providers:         []api.Provider{provider},
				RepositoryManager: mocks.NewMockRepositoryManager(ctrl),
				PackageManager:    mocks.NewMockPackageManager(ctrl),
			}
			if err := InstallSpecific(deps, []string{"amdgpu:latest", invalid}, false, false, true); err == nil {
				t.Fatal("expected invalid request to fail before repository or package calls")
			}
		})
	}
}

func TestExplicitMixedInstallPreparesSelectedProviders(t *testing.T) {
	ctrl := gomock.NewController(t)
	rm := mocks.NewMockRepositoryManager(ctrl)
	pm := mocks.NewMockPackageManager(ctrl)
	nv := mocks.NewMockProvider(ctrl)
	am := mocks.NewMockProvider(ctrl)
	unused := mocks.NewMockProvider(ctrl)
	unused.EXPECT().GetID().Return("unused").AnyTimes()
	providers := []api.Provider{nv, am, unused}
	prepared := rm.EXPECT().EnsureRepositoriesEnabled([]string{"BaseOS", "Supplementary", "Extensions"}).Return(nil)
	configured := pm.EXPECT().SetEnableRepos(gomock.Nil()).After(prepared)
	for i, p := range []*mocks.MockProvider{nv, am} {
		id := []string{"nvidia", "amdgpu"}[i]
		channels := [][]string{{"BaseOS", "Supplementary"}, {"BaseOS", "Extensions"}}[i]
		p.EXPECT().GetID().Return(id).AnyTimes()
		p.EXPECT().GetName().Return(id).AnyTimes()
		p.EXPECT().RequiredChannels().Return(channels).Times(1)
		driver := api.DriverID{ProviderID: id, Version: "1"}
		available := p.EXPECT().ListAvailable().After(configured).Return([]api.DriverID{driver}, nil)
		p.EXPECT().Install([]api.DriverID{driver}).After(available).Return([]string{id + "-pkg"}, nil)
	}
	pm.EXPECT().Install([]string{"nvidia-pkg", "amdgpu-pkg"}, true, false).Return(nil)
	deps := api.CoreDeps{Providers: providers, RepositoryManager: rm, PackageManager: pm}
	if err := InstallSpecific(deps, []string{"amdgpu:1", "nvidia:1"}, true, false, true); err != nil {
		t.Fatal(err)
	}
}

func TestAutoDetectPreparesRepositoriesAfterDetection(t *testing.T) {
	ctrl := gomock.NewController(t)
	rm := mocks.NewMockRepositoryManager(ctrl)
	pm := mocks.NewMockPackageManager(ctrl)
	selected := mocks.NewMockProvider(ctrl)
	absent := mocks.NewMockProvider(ctrl)
	selected.EXPECT().GetID().Return("third").AnyTimes()
	selected.EXPECT().GetName().Return("Third GPU").AnyTimes()
	absent.EXPECT().GetID().Return("absent").AnyTimes()
	detected := selected.EXPECT().DetectHardware().Return(true, nil)
	notDetected := absent.EXPECT().DetectHardware().Return(false, nil)
	selected.EXPECT().RequiredChannels().Return([]string{"ThirdChannel"})
	prepared := rm.EXPECT().EnsureRepositoriesEnabled([]string{"ThirdChannel"}).After(detected).After(notDetected).Return(nil)
	configured := pm.EXPECT().SetEnableRepos(gomock.Nil()).After(prepared)
	available := selected.EXPECT().ListAvailable().After(configured).Return([]api.DriverID{{ProviderID: "third", Version: "1"}}, nil)
	selected.EXPECT().Install([]api.DriverID{{ProviderID: "third", Version: "1"}}).After(available).Return([]string{"third-pkg"}, nil)
	pm.EXPECT().Install([]string{"third-pkg"}, false, false).Return(nil)
	deps := api.CoreDeps{Providers: []api.Provider{selected, absent}, RepositoryManager: rm, PackageManager: pm}
	if err := InstallAutoDetect(deps, false, false, false); err != nil {
		t.Fatal(err)
	}
}

func TestListClearsRepositoryOverrides(t *testing.T) {
	ctrl := gomock.NewController(t)
	rm := mocks.NewMockRepositoryManager(ctrl)
	pm := mocks.NewMockPackageManager(ctrl)
	provider := mocks.NewMockProvider(ctrl)
	provider.EXPECT().GetName().Return("Third GPU").AnyTimes()
	provider.EXPECT().RequiredChannels().Return([]string{"ThirdChannel"})
	resolved := rm.EXPECT().GetRepoIDs([]string{"ThirdChannel"}).Return(nil, nil)
	cleared := pm.EXPECT().SetEnableRepos(gomock.Nil()).After(resolved)
	provider.EXPECT().ListAvailable().After(cleared).Return(nil, nil)
	_, err := List(api.CoreDeps{Providers: []api.Provider{provider}, RepositoryManager: rm, PackageManager: pm}, false, true, false, false)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCompatibleListDetectionFailureNeedsNoRepositories(t *testing.T) {
	ctrl := gomock.NewController(t)
	provider := mocks.NewMockProvider(ctrl)
	provider.EXPECT().GetName().Return("Third GPU")
	provider.EXPECT().DetectHardware().Return(false, errors.New("detection failed"))
	result, err := List(api.CoreDeps{Providers: []api.Provider{provider}}, true, true, false, true)
	if err != nil || len(result) != 0 {
		t.Fatalf("List() = %v, %v", result, err)
	}
}
