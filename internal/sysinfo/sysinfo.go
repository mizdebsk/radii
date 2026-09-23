package sysinfo

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mizdebsk/radii/internal/log"
)

const (
	cloudInfoPath     = "/run/cloud-init/instance-data.json"
	osReleasePath     = "/etc/os-release"
	kernelReleasePath = "/proc/sys/kernel/osrelease"
)

type SysInfo struct {
	IsRhel        bool
	OsVersion     int
	Arch          string
	KernelVersion string
	KernelVariant string
	CloudProvider string
}

func DetectSysInfo() (SysInfo, error) {
	info, err := detectSysInfo(osReleasePath, kernelReleasePath)
	if err != nil {
		return SysInfo{}, err
	}
	info.CloudProvider = detectCloudProvider(cloudInfoPath)
	return info, nil
}

func detectSysInfo(osPath, kernelPath string) (SysInfo, error) {
	isRhel, osVersion := detectOs(osPath)
	kernel, err := detectKernel(kernelPath)
	if err != nil {
		return SysInfo{}, fmt.Errorf("unable to detect running kernel: %w", err)
	}
	return SysInfo{
		IsRhel:        isRhel,
		OsVersion:     osVersion,
		Arch:          kernel.arch,
		KernelVersion: kernel.version,
		KernelVariant: kernel.variant,
	}, nil
}

type kernelInfo struct {
	version string
	arch    string
	variant string
}

func detectKernel(path string) (kernelInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return kernelInfo{}, err
	}
	release := strings.TrimSpace(string(data))
	base, variant, hasVariant := strings.Cut(release, "+")
	archStart := strings.LastIndexByte(base, '.')
	if archStart < 0 || strings.ContainsAny(release, " \t\r\n") ||
		(hasVariant && (variant == "" || strings.Contains(variant, "+"))) {
		return kernelInfo{}, fmt.Errorf("invalid kernel release %q", release)
	}
	version, arch := base[:archStart], base[archStart+1:]
	upstream, revision, hasRevision := strings.Cut(version, "-")
	if !hasRevision || upstream == "" || revision == "" || arch == "" ||
		strings.ContainsAny(arch, "-+") {
		return kernelInfo{}, fmt.Errorf("invalid kernel release %q", release)
	}
	return kernelInfo{version: version, arch: arch, variant: variant}, nil
}

func detectOs(path string) (bool, int) {
	f, err := os.Open(path)
	if err != nil {
		log.Logf("unable to open %s for reading: %v", path, err)
		return false, 0
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Warnf("failed to close file %s: %v", path, err)
		}
	}()

	var isRhel bool
	var osVersion int

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "ID=") {
			val := strings.TrimPrefix(line, "ID=")
			val = strings.Trim(val, `"`)
			isRhel = val == "rhel"
		} else if strings.HasPrefix(line, "VERSION_ID=") {
			val := strings.TrimPrefix(line, "VERSION_ID=")
			val = strings.Trim(val, `"`)
			if idx := strings.IndexByte(val, '.'); idx >= 0 {
				val = val[:idx]
			}
			n, err := strconv.Atoi(val)
			if err != nil {
				log.Warnf("invalid VERSION_ID %q in %s: %v", val, path, err)
			}
			osVersion = n
		}
	}
	if err := scanner.Err(); err != nil {
		log.Warnf("error parsing %s: %v", path, err)
		return false, 0
	}
	return isRhel, osVersion
}

type cloudInfo struct {
	V1 struct {
		CloudName string `json:"cloud_name"`
	} `json:"v1"`
}

func detectCloudProvider(path string) string {
	cloudInfoJson, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Logf("cloud instance data absent: %s", path)
		} else {
			log.Warnf("failed to read cloud instance data file %s: %v", path, err)
		}
		return ""
	}

	var info cloudInfo
	if err := json.Unmarshal(cloudInfoJson, &info); err != nil {
		log.Warnf("failed to parse cloud instance data file %s: %v", path, err)
		return ""
	}

	return info.V1.CloudName
}
