package cli

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mizdebsk/radii/internal/api"
	"github.com/mizdebsk/radii/internal/mocks"
	"github.com/mizdebsk/radii/internal/rhsm"
	"github.com/mizdebsk/radii/internal/sysinfo"
)

func TestSkipSubscriptionsUsesExistingSources(t *testing.T) {
	for _, command := range []string{"list", "install", "dry-run"} {
		t.Run(command, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			executor := mocks.NewMockExecutor(ctrl)
			pm := mocks.NewMockPackageManager(ctrl)
			p := mocks.NewMockProvider(ctrl)
			p.EXPECT().GetID().Return("amdgpu").AnyTimes()
			p.EXPECT().GetName().Return("AMD GPU").AnyTimes()
			p.EXPECT().RequiredChannels().Return([]string{"UnconfiguredTestChannel"})
			configured := pm.EXPECT().SetEnableRepos(gomock.Nil())
			drivers := []api.DriverID{{ProviderID: "amdgpu", Version: "latest"}}
			p.EXPECT().ListAvailable().After(configured).Return(drivers, nil)
			args := []string{"radii", "--skip-subscriptions"}
			if command == "list" {
				args = append(args, "list")
				p.EXPECT().DetectHardware().Return(true, nil)
				p.EXPECT().ListInstalled().Return(nil, nil)
			} else {
				args = append(args, "install", "--force")
				if command == "dry-run" {
					args = append(args, "--dry-run")
				}
				args = append(args, "amdgpu:latest")
				p.EXPECT().Install(drivers).Return([]string{"kmod-amdgpu", "rocm-devel"}, nil)
				pm.EXPECT().Install([]string{"kmod-amdgpu", "rocm-devel"}, false, command == "dry-run").Return(nil)
			}
			deps := api.CoreDeps{Providers: []api.Provider{p}, PackageManager: pm,
				RepositoryManager: rhsm.NewRepositoryManager(executor, sysinfo.SysInfo{IsRhel: true, OsVersion: 10, Arch: "x86_64"})}
			if err := Execute(args, deps, "test"); err != nil {
				t.Fatal(err)
			}
		})
	}
}
