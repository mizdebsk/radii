package sysinfo

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectSysInfo(t *testing.T) {
	tests := []struct {
		arch    string
		variant string
	}{
		{arch: "x86_64"},
		{arch: "aarch64"},
		{arch: "aarch64", variant: "64k"},
		{arch: "ppc64le"},
		{arch: "s390x"},
	}
	for _, tt := range tests {
		t.Run(tt.arch+"/"+tt.variant, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "osrelease")
			release := "6.12.0-211.51.1.el10_2." + tt.arch
			if tt.variant != "" {
				release += "+" + tt.variant
			}
			if err := os.WriteFile(path, []byte(release+"\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			got, err := detectSysInfo("testdata/os-release-rhel-10.1", path)
			if err != nil {
				t.Fatal(err)
			}
			want := SysInfo{
				IsRhel:        true,
				OsVersion:     10,
				Arch:          tt.arch,
				KernelVersion: "6.12.0-211.51.1.el10_2",
				KernelVariant: tt.variant,
			}
			if got != want {
				t.Fatalf("detectSysInfo() = %+v, want %+v", got, want)
			}
		})
	}
}

func TestDetectSysInfoKernelUnavailable(t *testing.T) {
	for _, name := range []string{"missing", "invalid", "empty", "directory"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "osrelease")
			if name == "invalid" {
				if err := os.WriteFile(path, []byte("invalid\n"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			if name == "empty" {
				if err := os.WriteFile(path, nil, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			if name == "directory" {
				if err := os.Mkdir(path, 0o700); err != nil {
					t.Fatal(err)
				}
			}
			got, err := detectSysInfo("testdata/os-release-rhel-10.1", path)
			if err == nil || !strings.Contains(err.Error(), "unable to detect running kernel") || got != (SysInfo{}) {
				t.Fatalf("detectSysInfo() = %+v, %v; want empty result and kernel detection error", got, err)
			}
			if name == "missing" && !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("expected missing-file error, got %v", err)
			}
		})
	}
}

func TestDetectRhelVersion(t *testing.T) {

	tests := []struct {
		name      string
		path      string
		isRhel    bool
		osVersion int
	}{
		{
			name:      "RHEL 10.1",
			path:      "testdata/os-release-rhel-10.1",
			isRhel:    true,
			osVersion: 10,
		},
		{
			name:      "Fedora ELN 44",
			path:      "testdata/os-release-eln",
			isRhel:    false,
			osVersion: 44,
		},
		{
			name:      "Fedora Linux 43",
			path:      "testdata/os-release-fedora",
			isRhel:    false,
			osVersion: 43,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isRhel, rhelVersion := detectOs(tt.path)
			if isRhel != tt.isRhel || rhelVersion != tt.osVersion {
				t.Fatalf("detectOs(%q) = (%v, %v), want (%v, %v)", tt.path, isRhel, rhelVersion, tt.isRhel, tt.osVersion)
			}
		})
	}
}

func TestDetectCloud(t *testing.T) {

	tests := []struct {
		name          string
		path          string
		cloudProvider string
	}{
		{
			name:          "AWS",
			path:          "testdata/cloud-aws.json",
			cloudProvider: "aws",
		},
		{
			name:          "Azure",
			path:          "testdata/cloud-azure.json",
			cloudProvider: "azure",
		},
		{
			name:          "GCE",
			path:          "testdata/cloud-gce.json",
			cloudProvider: "gce",
		},
		{
			name:          "CloudDataNonExistent",
			path:          "testdata/this-file-does-not-exist.json",
			cloudProvider: "",
		},
		{
			name:          "CloudInvalidFormat",
			path:          "testdata/cloud-invalid.json",
			cloudProvider: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cloudProvider := detectCloudProvider(tt.path)
			if cloudProvider != tt.cloudProvider {
				t.Fatalf("detectCloudProvider(%q) = %q, %q", tt.path, cloudProvider, tt.cloudProvider)
			}
		})
	}
}
