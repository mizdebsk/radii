package core

import (
	"strings"
	"testing"

	"github.com/mizdebsk/radii/internal/api"
	"github.com/mizdebsk/radii/internal/sysinfo"
)

func TestResolveKernel(t *testing.T) {
	running := sysinfo.SysInfo{Arch: "aarch64", KernelVersion: "6.12.0-124.15.1.el10_1"}
	running64k := running
	running64k.KernelVariant = "64k"
	defaultVariant, fourK, emptyVariant, sixtyFourK := "default", "4k", "", "64k"
	version := "6.12.0-211.51.1.el10_2"
	tests := []struct {
		name    string
		system  sysinfo.SysInfo
		options api.KernelOptions
		want    api.KernelTarget
	}{
		{
			name: "default running kernel without version pin", system: running,
			want: api.KernelTarget{Arch: "aarch64"},
		},
		{
			name: "64k running kernel without version pin", system: running64k,
			want: api.KernelTarget{Arch: "aarch64", Variant: "64k"},
		},
		{
			name: "variant only", system: running,
			options: api.KernelOptions{Variant: &sixtyFourK},
			want:    api.KernelTarget{Arch: "aarch64", Variant: "64k"},
		},
		{
			name: "explicit default overrides running 64k", system: running64k,
			options: api.KernelOptions{Variant: &defaultVariant},
			want:    api.KernelTarget{Arch: "aarch64"},
		},
		{
			name: "4k alias overrides running 64k", system: running64k,
			options: api.KernelOptions{Variant: &fourK},
			want:    api.KernelTarget{Arch: "aarch64"},
		},
		{
			name: "explicit empty overrides running 64k", system: running64k,
			options: api.KernelOptions{Variant: &emptyVariant},
			want:    api.KernelTarget{Arch: "aarch64"},
		},
		{
			name: "version only follows running variant", system: running64k,
			options: api.KernelOptions{Version: version},
			want:    api.KernelTarget{Version: version, Arch: "aarch64", Variant: "64k"},
		},
		{
			name: "version and variant", system: running64k,
			options: api.KernelOptions{Version: version, Variant: &fourK},
			want:    api.KernelTarget{Version: version, Arch: "aarch64"},
		},
		{
			name: "matching architecture suffix", system: running64k,
			options: api.KernelOptions{Version: version + ".aarch64"},
			want:    api.KernelTarget{Version: version, Arch: "aarch64", Variant: "64k"},
		},
		{
			name: "uname release selects embedded variant", system: running,
			options: api.KernelOptions{Version: version + ".aarch64+64k"},
			want:    api.KernelTarget{Version: version, Arch: "aarch64", Variant: "64k"},
		},
		{
			name: "matching explicit and embedded variants", system: running,
			options: api.KernelOptions{Version: version + ".aarch64+64k", Variant: &sixtyFourK},
			want:    api.KernelTarget{Version: version, Arch: "aarch64", Variant: "64k"},
		},
		{
			name: "kernel without distribution tag", system: running,
			options: api.KernelOptions{Version: "6.12.0-211.51.1"},
			want:    api.KernelTarget{Version: "6.12.0-211.51.1", Arch: "aarch64"},
		},
		{
			name: "Fedora release", system: sysinfo.SysInfo{Arch: "x86_64"},
			options: api.KernelOptions{Version: "7.1.3-100.fc43.x86_64"},
			want:    api.KernelTarget{Version: "7.1.3-100.fc43", Arch: "x86_64"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveKernel(tt.system, tt.options)
			if err != nil || got != tt.want {
				t.Fatalf("resolveKernel() = %+v, %v; want %+v, nil", got, err, tt.want)
			}
		})
	}
}

func TestResolveKernelRejectsInvalidOptions(t *testing.T) {
	fourK, unknown := "4k", "16k"
	tests := []struct {
		name    string
		arch    string
		options api.KernelOptions
		wantErr string
	}{
		{name: "missing architecture", wantErr: "architecture"},
		{name: "invalid variant", arch: "aarch64", options: api.KernelOptions{Variant: &unknown}, wantErr: "unsupported kernel variant"},
		{name: "conflicting variant", arch: "aarch64", options: api.KernelOptions{Version: "6.12.0-211.el10.aarch64+64k", Variant: &fourK}, wantErr: "conflicts"},
		{name: "empty embedded variant", arch: "aarch64", options: api.KernelOptions{Version: "6.12.0-211.el10.aarch64+"}, wantErr: "empty variant"},
		{name: "unknown embedded variant", arch: "aarch64", options: api.KernelOptions{Version: "6.12.0-211.el10.aarch64+debug"}, wantErr: "unsupported kernel variant"},
		{name: "variant without version", arch: "aarch64", options: api.KernelOptions{Version: "+64k"}, wantErr: "invalid kernel version"},
		{name: "mismatched architecture", arch: "aarch64", options: api.KernelOptions{Version: "6.12.0-211.el10.x86_64"}, wantErr: "invalid kernel version"},
		{name: "unknown architecture", arch: "aarch64", options: api.KernelOptions{Version: "6.12.0-211.el10.mips"}, wantErr: "invalid kernel version"},
		{name: "missing release", arch: "aarch64", options: api.KernelOptions{Version: "6.12.0"}, wantErr: "invalid kernel version"},
		{name: "glob", arch: "aarch64", options: api.KernelOptions{Version: "6.12.0-*"}, wantErr: "invalid kernel version"},
		{name: "extra argument", arch: "aarch64", options: api.KernelOptions{Version: "6.12.0-211.el10 --allowerasing"}, wantErr: "invalid kernel version"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveKernel(sysinfo.SysInfo{Arch: tt.arch}, tt.options)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) || got != (api.KernelTarget{}) {
				t.Fatalf("resolveKernel() = %+v, %v; want empty target and error containing %q", got, err, tt.wantErr)
			}
		})
	}
}
