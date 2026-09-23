package core

import (
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mizdebsk/radii/internal/api"
	"github.com/mizdebsk/radii/internal/mocks"
	"github.com/mizdebsk/radii/internal/sysinfo"
)

func TestInstallPassesKernelTarget(t *testing.T) {
	defaultVariant := "default"
	for _, mode := range []string{"specific", "automatic"} {
		for _, tt := range []struct {
			name    string
			options api.KernelOptions
			want    api.KernelTarget
		}{
			{"running variant", api.KernelOptions{}, api.KernelTarget{Variant: "64k", Arch: "aarch64"}},
			{"variant only", api.KernelOptions{Variant: &defaultVariant}, api.KernelTarget{Arch: "aarch64"}},
			{"version only", api.KernelOptions{Version: "6.12.0-211.51.1.el10_2"}, api.KernelTarget{Version: "6.12.0-211.51.1.el10_2", Variant: "64k", Arch: "aarch64"}},
			{"version and variant", api.KernelOptions{Version: "6.12.0-211.51.1.el10_2", Variant: &defaultVariant}, api.KernelTarget{Version: "6.12.0-211.51.1.el10_2", Arch: "aarch64"}},
		} {
			t.Run(mode+"/"+tt.name, func(t *testing.T) {
				ctrl := gomock.NewController(t)
				provider := mocks.NewMockProvider(ctrl)
				pm := mocks.NewMockPackageManager(ctrl)
				rm := mocks.NewMockRepositoryManager(ctrl)
				drivers := []api.DriverID{{ProviderID: "nvidia", Version: "580.178.04"}}
				provider.EXPECT().GetID().Return("nvidia").AnyTimes()
				provider.EXPECT().GetName().Return("NVIDIA").AnyTimes()
				provider.EXPECT().DetectHardware().Return(true, nil)
				provider.EXPECT().ListAvailable().Return(drivers, nil)
				provider.EXPECT().RequiredChannels().Return([]string{"TestChannel"})
				rm.EXPECT().GetRepoIDs([]string{"TestChannel"}).Return([]string{"test-repo"}, nil)
				pm.EXPECT().SetEnableRepos([]string{"test-repo"})
				provider.EXPECT().Install(drivers, tt.want).Return([]string{"selected-kmod"}, nil)
				pm.EXPECT().Install([]string{"selected-kmod"}, true, true).Return(nil)
				deps := api.CoreDeps{
					SystemInfo: sysinfo.SysInfo{
						Arch: "aarch64", KernelVersion: "6.12.0-211.50.1.el10_2", KernelVariant: "64k",
					},
					Providers: []api.Provider{provider}, PackageManager: pm, RepositoryManager: rm,
				}
				var err error
				if mode == "specific" {
					err = InstallSpecific(deps, []string{"nvidia:580.178.04"}, true, true, false, tt.options)
				} else {
					err = InstallAutoDetect(deps, true, true, false, tt.options)
				}
				if err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestInstallRejectsInvalidKernelBeforeSideEffects(t *testing.T) {
	for _, mode := range []string{"specific", "automatic"} {
		t.Run(mode, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			deps := api.CoreDeps{
				SystemInfo:        sysinfo.SysInfo{Arch: "aarch64"},
				Providers:         []api.Provider{mocks.NewMockProvider(ctrl)},
				PackageManager:    mocks.NewMockPackageManager(ctrl),
				RepositoryManager: mocks.NewMockRepositoryManager(ctrl),
			}
			options := api.KernelOptions{Version: "6.12.0-211.51.1.el10_2.x86_64"}
			var err error
			if mode == "specific" {
				err = InstallSpecific(deps, []string{"nvidia:580.178.04"}, false, false, false, options)
			} else {
				err = InstallAutoDetect(deps, false, false, false, options)
			}
			if err == nil || !strings.Contains(err.Error(), "invalid kernel version") {
				t.Fatalf("expected kernel validation error, got %v", err)
			}
		})
	}
}
