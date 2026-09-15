package rhsm

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/mizdebsk/radii/internal/sysinfo"
)

func TestCloudSubscriptionHelp(t *testing.T) {
	tests := []struct {
		name   string
		cloud  string
		rhel   int
		match1 string
		match2 string
	}{
		{
			name:   "AWS-RHEL9",
			cloud:  "aws",
			rhel:   9,
			match1: "Amazon Web Services",
			match2: "subscription-manager register",
		},
		{
			name:   "AWS-RHEL10",
			cloud:  "aws",
			rhel:   10,
			match1: "Amazon Web Services",
			match2: "rhc connect",
		},
		{
			name:   "Azure-RHEL9",
			cloud:  "azure",
			rhel:   9,
			match1: "Microsoft Azure",
			match2: "subscription-manager register",
		},
		{
			name:   "Azure-RHEL10",
			cloud:  "azure",
			rhel:   10,
			match1: "Microsoft Azure",
			match2: "rhc connect",
		},
		{
			name:   "GCE-RHEL9",
			cloud:  "gce",
			rhel:   9,
			match1: "Google Cloud",
			match2: "subscription-manager register",
		},
		{
			name:   "GCE-RHEL10",
			cloud:  "gce",
			rhel:   10,
			match1: "Google Cloud",
			match2: "rhc connect",
		},
		{
			name:   "OpenStack-RHEL5",
			cloud:  "openstack",
			rhel:   5,
			match1: "Please ensure that this system is registered",
			match2: "and has access to the required Red Hat repositories",
		},
		{
			name:   "OpenStack-RHEL9",
			cloud:  "openstack",
			rhel:   9,
			match1: "Please ensure that this system is registered",
			match2: "subscription-manager register",
		},
		{
			name:   "OpenStack-RHEL10",
			cloud:  "openstack",
			rhel:   10,
			match1: "Please ensure that this system is registered",
			match2: "rhc connect",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			si := sysinfo.SysInfo{
				IsRhel:        true,
				OsVersion:     tt.rhel,
				Arch:          "pdp11",
				CloudProvider: tt.cloud,
			}
			err := fmt.Errorf("KABOOM")
			detailedErr := getDetailedSubscriptionError(err, si)
			detailedMsg := detailedErr.Error()
			if !errors.Is(detailedErr, err) {
				t.Errorf("case %s: expected detailed error to wrap original error", tt.name)
			}
			for _, match := range []string{
				tt.match1,
				tt.match2,
				"https://docs.redhat.com/",
				"KABOOM",
			} {
				if !strings.Contains(detailedMsg, match) {
					t.Errorf("case %s: expected error message to have substring %q", tt.name, match)
				}
			}
		})
	}
}
