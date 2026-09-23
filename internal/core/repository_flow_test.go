package core

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mizdebsk/radii/internal/api"
	"github.com/mizdebsk/radii/internal/mocks"
	"github.com/mizdebsk/radii/internal/sysinfo"
)

func TestInstallValidatesAllArgumentsBeforeRepositorySetup(t *testing.T) {
	for _, invalid := range []string{"invalid-format", "unknown:1"} {
		t.Run(invalid, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			provider := mocks.NewMockProvider(ctrl)
			provider.EXPECT().GetID().Return("amdgpu").AnyTimes()
			deps := api.CoreDeps{SystemInfo: sysinfo.SysInfo{Arch: "x86_64"},
				Providers:         []api.Provider{provider},
				RepositoryManager: mocks.NewMockRepositoryManager(ctrl),
				PackageManager:    mocks.NewMockPackageManager(ctrl),
			}
			if err := InstallSpecific(deps, []string{"amdgpu:latest", invalid}, false, false, true, api.KernelOptions{}); err == nil {
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
		p.EXPECT().Install([]api.DriverID{driver}, api.KernelTarget{Arch: "x86_64"}).After(available).Return([]string{id + "-pkg"}, nil)
	}
	pm.EXPECT().Install([]string{"nvidia-pkg", "amdgpu-pkg"}, true, false).Return(nil)
	deps := api.CoreDeps{SystemInfo: sysinfo.SysInfo{Arch: "x86_64"}, Providers: providers, RepositoryManager: rm, PackageManager: pm}
	if err := InstallSpecific(deps, []string{"amdgpu:1", "nvidia:1"}, true, false, true, api.KernelOptions{}); err != nil {
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
	selected.EXPECT().Install([]api.DriverID{{ProviderID: "third", Version: "1"}}, api.KernelTarget{Arch: "x86_64"}).After(available).Return([]string{"third-pkg"}, nil)
	pm.EXPECT().Install([]string{"third-pkg"}, false, false).Return(nil)
	deps := api.CoreDeps{SystemInfo: sysinfo.SysInfo{Arch: "x86_64"}, Providers: []api.Provider{selected, absent}, RepositoryManager: rm, PackageManager: pm}
	if err := InstallAutoDetect(deps, false, false, false, api.KernelOptions{}); err != nil {
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
	_, err := List(api.CoreDeps{SystemInfo: sysinfo.SysInfo{Arch: "x86_64"}, Providers: []api.Provider{provider}, RepositoryManager: rm, PackageManager: pm}, false, true, false, false)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCompatibleListDetectionFailureNeedsNoRepositories(t *testing.T) {
	ctrl := gomock.NewController(t)
	provider := mocks.NewMockProvider(ctrl)
	provider.EXPECT().GetName().Return("Third GPU")
	provider.EXPECT().DetectHardware().Return(false, errors.New("detection failed"))
	result, err := List(api.CoreDeps{SystemInfo: sysinfo.SysInfo{Arch: "x86_64"}, Providers: []api.Provider{provider}}, true, true, false, true)
	if err != nil || len(result) != 0 {
		t.Fatalf("List() = %v, %v", result, err)
	}
}

func TestExplicitInstallRejectsHardwareBeforeRepositorySetup(t *testing.T) {
	for _, dryRun := range []bool{false, true} {
		name := "install"
		if dryRun {
			name = "dry run"
		}
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			nv := mocks.NewMockProvider(ctrl)
			am := mocks.NewMockProvider(ctrl)
			nv.EXPECT().GetID().Return("nvidia").AnyTimes()
			am.EXPECT().GetID().Return("amdgpu").AnyTimes()
			nv.EXPECT().GetName().Return("NVIDIA").AnyTimes()
			am.EXPECT().GetName().Return("AMD GPU").AnyTimes()
			nv.EXPECT().DetectHardware().Return(true, nil)
			am.EXPECT().DetectHardware().Return(false, nil)
			deps := api.CoreDeps{
				SystemInfo:        sysinfo.SysInfo{Arch: "x86_64"},
				Providers:         []api.Provider{nv, am},
				RepositoryManager: mocks.NewMockRepositoryManager(ctrl),
				PackageManager:    mocks.NewMockPackageManager(ctrl),
			}
			err := InstallSpecific(deps, []string{"nvidia", "amdgpu"}, false, dryRun, false, api.KernelOptions{})
			if err == nil || err.Error() != "no compatible AMD GPU hardware found" {
				t.Fatalf("error = %v, want incompatible AMD hardware error", err)
			}
		})
	}
}

func TestExplicitInstallDetectsProviderBeforeRepositories(t *testing.T) {
	for _, tt := range []struct {
		name      string
		force     bool
		detectErr error
	}{
		{name: "compatible"},
		{name: "detection error", detectErr: errors.New("detection failed")},
		{name: "forced", force: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			provider := mocks.NewMockProvider(ctrl)
			rm := mocks.NewMockRepositoryManager(ctrl)
			pm := mocks.NewMockPackageManager(ctrl)
			provider.EXPECT().GetID().Return("nvidia").AnyTimes()
			provider.EXPECT().GetName().Return("NVIDIA").AnyTimes()
			channels := provider.EXPECT().RequiredChannels().Return([]string{"BaseOS"})
			if !tt.force {
				detected := provider.EXPECT().DetectHardware().Return(tt.detectErr == nil, tt.detectErr)
				channels.After(detected)
			}
			prepared := rm.EXPECT().EnsureRepositoriesEnabled([]string{"BaseOS"}).After(channels).Return(nil)
			configured := pm.EXPECT().SetEnableRepos(gomock.Nil()).After(prepared)
			drivers := []api.DriverID{{ProviderID: "nvidia", Version: "580.100"}, {ProviderID: "nvidia", Version: "580.9"}}
			provider.EXPECT().ListAvailable().After(configured).Return(drivers, nil).Times(2)
			provider.EXPECT().Install(drivers, api.KernelTarget{Arch: "x86_64"}).Return([]string{"selected-packages"}, nil)
			pm.EXPECT().Install([]string{"selected-packages"}, false, false).Return(nil)
			deps := api.CoreDeps{
				SystemInfo: sysinfo.SysInfo{Arch: "x86_64"}, Providers: []api.Provider{provider},
				RepositoryManager: rm, PackageManager: pm,
			}
			if err := InstallSpecific(deps, []string{"nvidia:580.100", "nvidia:580.9"}, false, false, tt.force, api.KernelOptions{}); err != nil {
				t.Fatal(err)
			}
		})
	}
}
