package intel

import (
	"errors"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mizdebsk/radii/internal/api"
	"github.com/mizdebsk/radii/internal/core"
	"github.com/mizdebsk/radii/internal/mocks"
	"github.com/mizdebsk/radii/internal/sysinfo"
)

func TestInstallOpenVINO(t *testing.T) {
	for _, id := range []string{"openvino", "openvino:default", "openvino:latest"} {
		t.Run(id, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			pm := mocks.NewMockPackageManager(ctrl)
			rm := mocks.NewMockRepositoryManager(ctrl)
			gomock.InOrder(
				rm.EXPECT().GetRepoIDs([]string{"BaseOS", "AppStream", "Extensions"}).Return([]string{"extensions"}, nil),
				pm.EXPECT().SetEnableRepos([]string{"extensions"}),
				pm.EXPECT().ListAvailablePackages().Return([]api.PackageInfo{{Name: "openvino"}}, nil),
				pm.EXPECT().Install([]string{"openvino"}, false, true).Return(nil),
			)
			deps := api.CoreDeps{
				SystemInfo: sysinfo.SysInfo{Arch: "x86_64"},
				Providers:  []api.Provider{NewProvider(pm)}, PackageManager: pm, RepositoryManager: rm,
			}
			if err := core.InstallSpecific(deps, []string{id}, false, true, true, api.KernelOptions{}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestInstallRejectsUnsupportedRequests(t *testing.T) {
	for _, tt := range []struct {
		name    string
		version string
		kernel  api.KernelTarget
	}{
		{name: "unknown variant", version: "unknown"},
		{name: "specific kernel", version: "latest", kernel: api.KernelTarget{Version: "6.12.0-211.51.1.el10_2"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			packages, err := NewProvider(nil).Install([]api.DriverID{{ProviderID: "openvino", Version: tt.version}}, tt.kernel)
			if err == nil || len(packages) != 0 {
				t.Fatalf("Install() = %v, %v, want no packages and an error", packages, err)
			}
		})
	}
}

func TestListPackageStates(t *testing.T) {
	queryErr := errors.New("package query failed")
	for _, tt := range []struct {
		name     string
		packages []api.PackageInfo
		err      error
		want     []api.DriverID
	}{
		{name: "absent", want: []api.DriverID{}},
		{name: "compute runtime only", packages: []api.PackageInfo{{Name: "intel-compute-runtime"}}, want: []api.DriverID{}},
		{name: "present", packages: []api.PackageInfo{{Name: "openvino"}}, want: []api.DriverID{{ProviderID: "openvino", Version: "latest"}}},
		{name: "query failure", err: queryErr},
	} {
		for _, installed := range []bool{false, true} {
			mode := "available"
			if installed {
				mode = "installed"
			}
			t.Run(mode+"/"+tt.name, func(t *testing.T) {
				pm := mocks.NewMockPackageManager(gomock.NewController(t))
				list := NewProvider(pm).ListAvailable
				if installed {
					pm.EXPECT().ListInstalledPackages().Return(tt.packages, tt.err)
					list = NewProvider(pm).ListInstalled
				} else {
					pm.EXPECT().ListAvailablePackages().Return(tt.packages, tt.err)
				}
				got, err := list()
				if !errors.Is(err, tt.err) || !reflect.DeepEqual(got, tt.want) {
					t.Fatalf("list() = %v, %v, want %v, %v", got, err, tt.want, tt.err)
				}
			})
		}
	}
}

func TestRemovePackageStates(t *testing.T) {
	queryErr := errors.New("package query failed")
	for _, tt := range []struct {
		name     string
		version  string
		packages []api.PackageInfo
		err      error
		want     []string
	}{
		{name: "absent", packages: []api.PackageInfo{{Name: "intel-compute-runtime"}}, want: []string{}},
		{name: "bare ID", packages: []api.PackageInfo{{Name: "openvino"}, {Name: "intel-compute-runtime"}}, want: []string{"openvino"}},
		{name: "latest", version: "latest", packages: []api.PackageInfo{{Name: "openvino"}}, want: []string{"openvino"}},
		{name: "query failure", err: queryErr},
	} {
		t.Run(tt.name, func(t *testing.T) {
			pm := mocks.NewMockPackageManager(gomock.NewController(t))
			pm.EXPECT().ListInstalledPackages().Return(tt.packages, tt.err)
			got, err := NewProvider(pm).Remove([]api.DriverID{{ProviderID: "openvino", Version: tt.version}})
			if !errors.Is(err, tt.err) || !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Remove() = %v, %v, want %v, %v", got, err, tt.want, tt.err)
			}
		})
	}
	if _, err := NewProvider(nil).Remove([]api.DriverID{{ProviderID: "openvino", Version: "unknown"}}); err == nil {
		t.Fatal("Remove() accepted an unknown variant")
	}
}
