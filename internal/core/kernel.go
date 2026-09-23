package core

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mizdebsk/radii/internal/api"
	"github.com/mizdebsk/radii/internal/sysinfo"
)

var kernelVersionPattern = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+-[0-9]+(?:\.[0-9]+)*(?:\.(?:el|fc)[0-9]+(?:_[0-9]+)?)?$`)

func resolveKernel(system sysinfo.SysInfo, options api.KernelOptions) (api.KernelTarget, error) {
	if system.Arch == "" {
		return api.KernelTarget{}, fmt.Errorf("unable to determine kernel architecture")
	}
	kernel := api.KernelTarget{Arch: system.Arch, Variant: system.KernelVariant}
	version, embeddedVariant, hasVariant := strings.Cut(options.Version, "+")
	if hasVariant {
		if embeddedVariant == "" {
			return api.KernelTarget{}, fmt.Errorf("empty variant in kernel version %q", options.Version)
		}
		var err error
		kernel.Variant, err = normalizeKernelVariant(embeddedVariant)
		if err != nil {
			return api.KernelTarget{}, err
		}
	}
	if options.Variant != nil {
		variant, err := normalizeKernelVariant(*options.Variant)
		if err != nil {
			return api.KernelTarget{}, err
		}
		if hasVariant && variant != kernel.Variant {
			return api.KernelTarget{}, fmt.Errorf("kernel version %q conflicts with variant %q", options.Version, *options.Variant)
		}
		kernel.Variant = variant
	}
	if options.Version == "" {
		return kernel, nil
	}
	version = strings.TrimSuffix(version, "."+system.Arch)
	if !kernelVersionPattern.MatchString(version) {
		return api.KernelTarget{}, fmt.Errorf("invalid kernel version %q (expected version-release[.arch][+variant])", options.Version)
	}
	kernel.Version = version
	return kernel, nil
}

func normalizeKernelVariant(variant string) (string, error) {
	switch variant {
	case "", "4k", "default":
		return "", nil
	case "64k":
		return "64k", nil
	default:
		return "", fmt.Errorf("unsupported kernel variant %q (expected default, 4k, or 64k)", variant)
	}
}
