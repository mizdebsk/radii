package sysinfo

import (
	"testing"
)

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
