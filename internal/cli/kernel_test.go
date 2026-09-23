package cli

import (
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mizdebsk/radii/internal/api"
	"github.com/mizdebsk/radii/internal/mocks"
	"github.com/mizdebsk/radii/internal/sysinfo"
)

func TestInstallKernelOptions(t *testing.T) {
	for _, automatic := range []bool{false, true} {
		mode := "explicit"
		if automatic {
			mode = "automatic"
		}
		for _, tt := range []struct {
			name    string
			args    []string
			version string
			variant string
		}{
			{name: "omitted", variant: "64k"},
			{name: "variant only", args: []string{"--kernel-variant", "default"}},
			{name: "4k alias", args: []string{"--kernel-variant=4k"}},
			{name: "explicit empty", args: []string{"--kernel-variant="}},
			{name: "64k", args: []string{"--kernel-variant=64k"}, variant: "64k"},
			{name: "version only", args: []string{"--kernel", "6.12.0-211.51.1.el10_2.aarch64"}, version: "6.12.0-211.51.1.el10_2", variant: "64k"},
			{name: "short kernel option", args: []string{"-K", "6.12.0-211.51.1.el10_2.aarch64"}, version: "6.12.0-211.51.1.el10_2", variant: "64k"},
			{name: "both", args: []string{"--kernel=6.12.0-211.51.1.el10_2", "--kernel-variant=4k"}, version: "6.12.0-211.51.1.el10_2"},
			{name: "embedded variant", args: []string{"--kernel=6.12.0-211.51.1.el10_2.aarch64+64k"}, version: "6.12.0-211.51.1.el10_2", variant: "64k"},
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
				target := api.KernelTarget{Version: tt.version, Variant: tt.variant, Arch: "aarch64"}
				provider.EXPECT().Install(drivers, target).Return([]string{"selected-kmod"}, nil)
				pm.EXPECT().Install([]string{"selected-kmod"}, true, true).Return(nil)
				deps := api.CoreDeps{
					SystemInfo: sysinfo.SysInfo{Arch: "aarch64", KernelVersion: "6.12.0-211.50.1.el10_2", KernelVariant: "64k"},
					Providers:  []api.Provider{provider}, PackageManager: pm, RepositoryManager: rm,
				}
				args := append([]string{"radii", "install", "--batch", "--dry-run"}, tt.args...)
				if automatic {
					args = append(args, "--auto-detect")
				} else {
					args = append(args, "nvidia:580.178.04")
				}
				if err := Execute(args, deps, "test"); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestInstallRejectsInvalidKernelOptions(t *testing.T) {
	for _, tt := range []struct {
		name string
		args []string
		want string
	}{
		{"missing version", []string{"--auto-detect", "--kernel"}, "flag needs an argument"},
		{"missing short kernel value", []string{"--auto-detect", "-K"}, "flag needs an argument"},
		{"missing variant", []string{"--auto-detect", "--kernel-variant"}, "flag needs an argument"},
		{"unknown variant", []string{"--kernel-variant=16k", "nvidia"}, "unsupported kernel variant"},
		{"invalid version", []string{"--auto-detect", "--kernel=invalid"}, "invalid kernel version"},
		{"conflicting variant", []string{"--kernel=6.12.0-211.51.1.el10_2.aarch64+64k", "--kernel-variant=4k", "nvidia"}, "conflicts"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			deps := api.CoreDeps{
				SystemInfo:        sysinfo.SysInfo{Arch: "aarch64"},
				Providers:         []api.Provider{mocks.NewMockProvider(ctrl)},
				PackageManager:    mocks.NewMockPackageManager(ctrl),
				RepositoryManager: mocks.NewMockRepositoryManager(ctrl),
			}
			err := Execute(append([]string{"radii", "install"}, tt.args...), deps, "test")
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want containing %q", err, tt.want)
			}
		})
	}
}
