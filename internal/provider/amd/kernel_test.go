package amd

import (
	"reflect"
	"strings"
	"testing"

	"github.com/mizdebsk/radii/internal/api"
)

func TestInstallKernelTarget(t *testing.T) {
	for _, variant := range []string{"", "64k", "debug"} {
		t.Run("variant="+variant, func(t *testing.T) {
			for _, version := range []string{"", "6.12.0-211.51.1.el10_2"} {
				t.Run("version="+version, func(t *testing.T) {
					kernel := api.KernelTarget{Version: version, Variant: variant, Arch: "aarch64"}
					packages, err := NewProvider(nil).Install([]api.DriverID{{ProviderID: "amdgpu", Version: "latest"}}, kernel)
					if version != "" {
						if err == nil || !strings.Contains(err.Error(), "specific kernel version") {
							t.Fatalf("expected specific kernel rejection, got %v", err)
						}
						if len(packages) != 0 {
							t.Errorf("returned packages on failure: %v", packages)
						}
						return
					}
					if err != nil {
						t.Fatal(err)
					}
					want := []string{"kmod-amdgpu", "rocm-devel"}
					if !reflect.DeepEqual(packages, want) {
						t.Errorf("packages = %v, want %v", packages, want)
					}
				})
			}
		})
	}
}
