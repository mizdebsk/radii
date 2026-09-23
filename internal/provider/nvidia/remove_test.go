package nvidia

import (
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mizdebsk/radii/internal/api"
	"github.com/mizdebsk/radii/internal/mocks"
)

func TestRemoveKernelModulesForDriverVersion(t *testing.T) {
	for _, tt := range []struct {
		name      string
		installed []api.PackageInfo
		wantKmods []string
	}{
		{
			name: "multiple kernels and variants",
			installed: []api.PackageInfo{
				{Name: "kmod-nvidia-open-580.178.04-6.12.0-211.50.1", Version: "580.178.04", Release: "2.el10_2", Arch: "aarch64"},
				{Name: "kmod-nvidia-open-580.178.04-6.12.0-211.51.1", Version: "580.178.04", Release: "3.el10_2", Arch: "aarch64"},
				{Name: "kmod-64k-nvidia-open-580.178.04-6.12.0-211.50.1", Version: "580.178.04", Release: "2.el10_2", Arch: "aarch64"},
				{Name: "kmod-64k-nvidia-open-580.178.04-6.12.0-211.51.1", Version: "580.178.04", Release: "3.el10_2", Arch: "aarch64"},
			},
			wantKmods: []string{
				"kmod-nvidia-open-580.178.04-6.12.0-211.50.1-580.178.04-2.el10_2.aarch64",
				"kmod-nvidia-open-580.178.04-6.12.0-211.51.1-580.178.04-3.el10_2.aarch64",
				"kmod-64k-nvidia-open-580.178.04-6.12.0-211.50.1-580.178.04-2.el10_2.aarch64",
				"kmod-64k-nvidia-open-580.178.04-6.12.0-211.51.1-580.178.04-3.el10_2.aarch64",
			},
		},
		{name: "no matching kmods"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			installed := append([]api.PackageInfo{
				{Name: "nvidia-driver", Version: "580.178.04", Release: "3.el10_2", Arch: "aarch64"},
				{Name: "nvidia-driver", Version: "579.1", Release: "1.el10_2", Arch: "aarch64"},
				{Name: "kmod-nvidia-open-579.1-6.12.0-211.51.1", Version: "579.1", Release: "1.el10_2", Arch: "aarch64"},
				{Name: "kmod-64k-nvidia-open-579.1-6.12.0-211.51.1", Version: "579.1", Release: "1.el10_2", Arch: "aarch64"},
				{Name: "kernel-core", Version: "6.12.0", Release: "211.51.1.el10_2", Arch: "aarch64"},
				{Name: "kernel-64k-core", Version: "6.12.0", Release: "211.51.1.el10_2", Arch: "aarch64"},
				{Name: "unrelated", Version: "580.178.04", Release: "3.el10_2", Arch: "aarch64"},
				{Name: "kmod-nvidia-openish", Version: "580.178.04", Release: "3.el10_2", Arch: "aarch64"},
			}, tt.installed...)
			ctrl := gomock.NewController(t)
			pm := mocks.NewMockPackageManager(ctrl)
			pm.EXPECT().ListInstalledPackages().Return(installed, nil)
			got, err := NewProvider(pm).Remove([]api.DriverID{{ProviderID: "nvidia", Version: "580.178.04"}})
			if err != nil {
				t.Fatal(err)
			}
			want := append([]string{"nvidia-driver-580.178.04-3.el10_2.aarch64"}, tt.wantKmods...)
			want = append(want, "cublasmp", "cuda-compat", "cuda-toolkit", "cudnn", "dnf-plugin-nvidia", "libnccl-devel", "libnccl-static")
			if !reflect.DeepEqual(got, want) {
				t.Errorf("Remove() = %v, want %v", got, want)
			}
		})
	}
}
