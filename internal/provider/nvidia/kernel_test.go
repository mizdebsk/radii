package nvidia

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mizdebsk/radii/internal/api"
	"github.com/mizdebsk/radii/internal/mocks"
)

func TestInstallKernelTarget(t *testing.T) {
	const version = "580.178.04"
	const kernelVersion = "6.12.0-211.51.1.el10_2"
	const defaultName = "kmod-nvidia-open-580.178.04-6.12.0-211.51.1"
	const sixtyFourKName = "kmod-64k-nvidia-open-580.178.04-6.12.0-211.51.1"
	available := []api.PackageInfo{
		{Name: "nvidia-driver", Version: version, Release: "3.el10_2", Arch: "aarch64"},
		{Name: defaultName, Version: version, Release: "3.el10_2", Arch: "aarch64"},
		{Name: defaultName, Version: version, Release: "2.el10_2", Arch: "aarch64"},
		{Name: defaultName, Version: version, Release: "99.el10_2", Arch: "x86_64"},
		{Name: sixtyFourKName, Version: version, Release: "3.el10_2", Arch: "aarch64"},
		{Name: sixtyFourKName, Version: version, Release: "2.el10_2", Arch: "aarch64"},
		{Name: "kmod-nvidia-open-580.178.04-6.12.0-211.50.1", Version: version, Release: "99.el10_2", Arch: "aarch64"},
		{Name: "kmod-64k-nvidia-open-580.178.04-6.12.0-211.50.1", Version: version, Release: "99.el10_2", Arch: "aarch64"},
		{Name: "kmod-nvidia-open-579.1-6.12.0-211.51.1", Version: "579.1", Release: "99.el10_2", Arch: "aarch64"},
	}
	tests := []struct {
		name     string
		kernel   api.KernelTarget
		wantKmod string
		wantErr  string
	}{
		{
			name: "latest default", kernel: api.KernelTarget{Arch: "aarch64"},
			wantKmod: "kmod-nvidia-open",
		},
		{
			name: "latest 64k", kernel: api.KernelTarget{Arch: "aarch64", Variant: "64k"},
			wantKmod: "kmod-64k-nvidia-open",
		},
		{
			name: "specific default", kernel: api.KernelTarget{Arch: "aarch64", Version: kernelVersion},
			wantKmod: "kmod-nvidia-open-580.178.04-6.12.0-211.51.1-580.178.04-3.el10_2.aarch64",
		},
		{
			name: "specific 64k", kernel: api.KernelTarget{Arch: "aarch64", Version: kernelVersion, Variant: "64k"},
			wantKmod: "kmod-64k-nvidia-open-580.178.04-6.12.0-211.51.1-580.178.04-3.el10_2.aarch64",
		},
		{
			name: "kernel without distribution tag", kernel: api.KernelTarget{Arch: "aarch64", Version: "6.12.0-211.51.1"},
			wantKmod: "kmod-nvidia-open-580.178.04-6.12.0-211.51.1-580.178.04-3.el10_2.aarch64",
		},
		{
			name: "missing specific kernel", kernel: api.KernelTarget{Arch: "aarch64", Version: "6.12.0-999.el10_2"},
			wantErr: "not available",
		},
		{
			name: "no kmod for target architecture", kernel: api.KernelTarget{Arch: "ppc64le", Version: kernelVersion},
			wantErr: "not available",
		},
		{
			name: "unsupported variant", kernel: api.KernelTarget{Arch: "aarch64", Variant: "debug"},
			wantErr: "unsupported NVIDIA kernel variant",
		},
		{
			name: "64k on x86_64", kernel: api.KernelTarget{Arch: "x86_64", Variant: "64k"},
			wantErr: "require aarch64",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			pm := mocks.NewMockPackageManager(ctrl)
			pm.EXPECT().ListAvailablePackages().Return(available, nil).MinTimes(1)
			provider := NewProvider(pm)
			got, err := provider.Install([]api.DriverID{{ProviderID: "nvidia", Version: version}}, tt.kernel)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) || len(got) != 0 {
					t.Fatalf("Install() = %v, %v; want no packages and error containing %q", got, err, tt.wantErr)
				}
				return
			}
			want := []string{
				"nvidia-driver-580.178.04-3.el10_2.aarch64",
				"cublasmp", "cuda-compat", "cuda-toolkit", "cudnn",
				"dnf-plugin-nvidia", "libnccl-devel", "libnccl-static",
				tt.wantKmod,
			}
			if err != nil || !reflect.DeepEqual(got, want) {
				t.Fatalf("Install() = %v, %v; want %v, nil", got, err, want)
			}
		})
	}
}

func TestInstallGenericKmodWithoutNamedPackage(t *testing.T) {
	for _, variant := range []string{"", "64k"} {
		t.Run(variant, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			pm := mocks.NewMockPackageManager(ctrl)
			pm.EXPECT().ListAvailablePackages().Return([]api.PackageInfo{
				{Name: "nvidia-driver", Version: "580.178.04", Release: "3.el10_2", Arch: "aarch64"},
			}, nil).MinTimes(1)
			provider := NewProvider(pm)
			got, err := provider.Install([]api.DriverID{{ProviderID: "nvidia", Version: "580.178.04"}}, api.KernelTarget{Arch: "aarch64", Variant: variant})
			if err != nil {
				t.Fatal(err)
			}
			want := "kmod-nvidia-open"
			if variant == "64k" {
				want = "kmod-64k-nvidia-open"
			}
			count := 0
			for _, pkg := range got {
				if pkg == want {
					count++
				}
			}
			if count != 1 {
				t.Fatalf("Install() = %v; want exactly one %s", got, want)
			}
		})
	}
}

func TestInstallKernelTargetQueryFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	pm := mocks.NewMockPackageManager(ctrl)
	queryErr := errors.New("query failed")
	pm.EXPECT().ListAvailablePackages().Return(nil, queryErr)
	got, err := NewProvider(pm).Install([]api.DriverID{{ProviderID: "nvidia", Version: "580.178.04"}}, api.KernelTarget{Arch: "aarch64", Version: "6.12.0-211.51.1.el10_2"})
	if !errors.Is(err, queryErr) || len(got) != 0 {
		t.Fatalf("Install() = %v, %v; want no packages and query error", got, err)
	}
}
