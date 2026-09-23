package sysinfo

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestDetectKernel(t *testing.T) {
	tests := []struct {
		name    string
		release string
		want    kernelInfo
	}{
		{
			name:    "x86_64",
			release: "6.12.0-211.51.1.el10_2.x86_64\n",
			want:    kernelInfo{version: "6.12.0-211.51.1.el10_2", arch: "x86_64"},
		},
		{
			name:    "aarch64 default",
			release: "6.12.0-211.51.1.el10_2.aarch64\n",
			want:    kernelInfo{version: "6.12.0-211.51.1.el10_2", arch: "aarch64"},
		},
		{
			name:    "aarch64 64k",
			release: "6.12.0-211.51.1.el10_2.aarch64+64k\n",
			want:    kernelInfo{version: "6.12.0-211.51.1.el10_2", arch: "aarch64", variant: "64k"},
		},
		{
			name:    "RHEL 9 64k",
			release: "5.14.0-284.50.1.el9_2.aarch64+64k\n",
			want:    kernelInfo{version: "5.14.0-284.50.1.el9_2", arch: "aarch64", variant: "64k"},
		},
		{
			name:    "ppc64le default",
			release: "6.12.0-211.51.1.el10_2.ppc64le\n",
			want:    kernelInfo{version: "6.12.0-211.51.1.el10_2", arch: "ppc64le"},
		},
		{
			name:    "s390x",
			release: "6.12.0-211.51.1.el10_2.s390x\n",
			want:    kernelInfo{version: "6.12.0-211.51.1.el10_2", arch: "s390x"},
		},
		{
			name:    "Fedora without trailing newline",
			release: "7.1.3-100.fc43.x86_64",
			want:    kernelInfo{version: "7.1.3-100.fc43", arch: "x86_64"},
		},
		{
			name:    "debug variant",
			release: "6.12.0-211.51.1.el10_2.x86_64+debug\n",
			want:    kernelInfo{version: "6.12.0-211.51.1.el10_2", arch: "x86_64", variant: "debug"},
		},
		{
			name:    "surrounding whitespace",
			release: " \t6.12.0-211.51.1.el10_2.aarch64+64k\r\n",
			want:    kernelInfo{version: "6.12.0-211.51.1.el10_2", arch: "aarch64", variant: "64k"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "osrelease")
			if err := os.WriteFile(path, []byte(tt.release), 0o600); err != nil {
				t.Fatal(err)
			}
			got, err := detectKernel(path)
			if err != nil {
				t.Fatalf("detectKernel() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("detectKernel() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestDetectKernelInvalidRelease(t *testing.T) {
	releases := []string{
		"",
		"\n",
		"invalid",
		"6.12.0",
		"6.12.0-.aarch64",
		"-211.el10.aarch64",
		"6.12.0-211.el10.",
		"6.12.0-211.el10.aarch64+",
		"6.12.0-211.el10.aarch64+64k+debug",
		"6.12.0-211.el10.aarch64+64 k",
		"6.12.0-211.el10.aarch64\nextra",
		"6.8.0-45-generic",
	}
	for _, release := range releases {
		t.Run(release, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "osrelease")
			if err := os.WriteFile(path, []byte(release), 0o600); err != nil {
				t.Fatal(err)
			}
			got, err := detectKernel(path)
			if err == nil || got != (kernelInfo{}) {
				t.Fatalf("detectKernel() = %+v, %v; want empty result and error", got, err)
			}
		})
	}
}

func TestDetectKernelMissingFile(t *testing.T) {
	got, err := detectKernel(filepath.Join(t.TempDir(), "missing"))
	if !errors.Is(err, os.ErrNotExist) || got != (kernelInfo{}) {
		t.Fatalf("detectKernel() = %+v, %v; want empty result and missing-file error", got, err)
	}
}
