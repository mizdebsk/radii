package cli

import (
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mizdebsk/radii/internal/api"
	"github.com/mizdebsk/radii/internal/mocks"
	"github.com/mizdebsk/radii/internal/sysinfo"
)

func TestInstallDriverSelection(t *testing.T) {
	for _, tt := range []struct {
		name     string
		argument string
		provider string
		versions []string
		want     string
		kernel   string
	}{
		{"bare NVIDIA", "nvidia", "nvidia", []string{"580.100", "580.9"}, "580.100", ""},
		{"default NVIDIA", "nvidia:default", "nvidia", []string{"580.100", "580.9"}, "580.100", ""},
		{"latest NVIDIA", "nvidia:latest", "nvidia", []string{"580.9", "580.100", "580.20"}, "580.100", ""},
		{"explicit NVIDIA", "nvidia:580.9", "nvidia", []string{"580.100", "580.9"}, "580.9", ""},
		{"bare AMD", "amdgpu", "amdgpu", []string{"latest"}, "latest", ""},
		{"default AMD", "amdgpu:default", "amdgpu", []string{"latest"}, "latest", ""},
		{"latest AMD", "amdgpu:latest", "amdgpu", []string{"latest"}, "latest", ""},
		{"provider default is older", "nvidia:default", "nvidia", []string{"580.9", "580.100"}, "580.9", ""},
		{"bare uses older default", "nvidia", "nvidia", []string{"580.9", "580.100"}, "580.9", ""},
		{"default with kernel", "nvidia:default", "nvidia", []string{"580.100", "580.9"}, "580.100", "6.12.0-211.51.1.el10_2"},
		{"latest with kernel", "nvidia:latest", "nvidia", []string{"580.9", "580.100"}, "580.100", "6.12.0-211.51.1.el10_2"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			provider := mocks.NewMockProvider(ctrl)
			pm := mocks.NewMockPackageManager(ctrl)
			rm := mocks.NewMockRepositoryManager(ctrl)
			provider.EXPECT().GetID().Return(tt.provider).AnyTimes()
			provider.EXPECT().GetName().Return(tt.provider).AnyTimes()
			provider.EXPECT().RequiredChannels().Return([]string{"TestChannel"})
			var available []api.DriverID
			for _, version := range tt.versions {
				available = append(available, api.DriverID{ProviderID: tt.provider, Version: version})
			}
			gomock.InOrder(
				rm.EXPECT().GetRepoIDs([]string{"TestChannel"}).Return([]string{"test-repo"}, nil),
				pm.EXPECT().SetEnableRepos([]string{"test-repo"}),
				provider.EXPECT().ListAvailable().Return(available, nil),
			)
			provider.EXPECT().DetectHardware().Return(true, nil)
			target := api.KernelTarget{Arch: "aarch64", Variant: "64k", Version: tt.kernel}
			selected := []api.DriverID{{ProviderID: tt.provider, Version: tt.want}}
			provider.EXPECT().Install(selected, target).Return([]string{"selected-package"}, nil)
			pm.EXPECT().Install([]string{"selected-package"}, true, true).Return(nil)
			deps := api.CoreDeps{
				SystemInfo: sysinfo.SysInfo{Arch: "aarch64", KernelVariant: "64k"},
				Providers:  []api.Provider{provider}, PackageManager: pm, RepositoryManager: rm,
			}
			args := []string{"radii", "install", "--batch", "--dry-run"}
			if tt.kernel != "" {
				args = append(args, "--kernel", tt.kernel, "--kernel-variant", "64k")
			}
			args = append(args, tt.argument)
			if err := Execute(args, deps, "test"); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestInstallAliasesWithoutAvailableDrivers(t *testing.T) {
	for _, argument := range []string{"nvidia", "nvidia:default", "nvidia:latest"} {
		t.Run(argument, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			provider := mocks.NewMockProvider(ctrl)
			pm := mocks.NewMockPackageManager(ctrl)
			rm := mocks.NewMockRepositoryManager(ctrl)
			provider.EXPECT().GetID().Return("nvidia").AnyTimes()
			provider.EXPECT().GetName().Return("NVIDIA").AnyTimes()
			provider.EXPECT().RequiredChannels().Return(nil)
			rm.EXPECT().GetRepoIDs(gomock.Any()).Return(nil, nil)
			pm.EXPECT().SetEnableRepos(gomock.Nil())
			provider.EXPECT().ListAvailable().Return(nil, nil)
			deps := api.CoreDeps{
				SystemInfo: sysinfo.SysInfo{Arch: "aarch64"},
				Providers:  []api.Provider{provider}, PackageManager: pm, RepositoryManager: rm,
			}
			err := Execute([]string{"radii", "install", "--dry-run", argument}, deps, "test")
			if err == nil || err.Error() != "no NVIDIA drivers available" {
				t.Fatalf("error = %v, want no NVIDIA drivers available", err)
			}
		})
	}
}

func TestInstallRejectsInvalidDriverNames(t *testing.T) {
	for _, argument := range []string{"", "nvidia:", ":latest", "nvidia:latest:extra", "unknown", "unknown:latest"} {
		t.Run(argument, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			provider := mocks.NewMockProvider(ctrl)
			provider.EXPECT().GetID().Return("nvidia").AnyTimes()
			deps := api.CoreDeps{
				SystemInfo:     sysinfo.SysInfo{Arch: "aarch64"},
				Providers:      []api.Provider{provider},
				PackageManager: mocks.NewMockPackageManager(ctrl), RepositoryManager: mocks.NewMockRepositoryManager(ctrl),
			}
			want := "invalid driver ID format"
			if strings.HasPrefix(argument, "unknown") {
				want = "unknown provider"
			}
			err := Execute([]string{"radii", "install", argument}, deps, "test")
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("error = %v, want containing %q", err, want)
			}
		})
	}
}
