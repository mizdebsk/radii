package sysinfo

import (
	"bufio"
	"encoding/json"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/mizdebsk/radii/internal/log"
)

const osReleasePath = "/etc/os-release"
const cloudInfoPath = "/run/cloud-init/instance-data.json"

type SysInfo struct {
	IsRhel        bool
	OsVersion     int
	Arch          string
	CloudProvider string
}

func DetectSysInfo() SysInfo {
	arch := detectArch()
	isRhel, osVersion := detectOs(osReleasePath)
	cloudProvider := detectCloudProvider(cloudInfoPath)
	return SysInfo{
		IsRhel:        isRhel,
		OsVersion:     osVersion,
		Arch:          arch,
		CloudProvider: cloudProvider,
	}
}

func detectArch() string {
	switch runtime.GOARCH {
	case "amd64":
		return "x86_64"
	case "arm64":
		return "aarch64"
	default:
		// ppc64le, s390x, etc.
		return runtime.GOARCH
	}
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
