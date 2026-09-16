package rhsm

import (
	"fmt"

	"github.com/mizdebsk/radii/internal/sysinfo"
)

var genericHelp = map[int]string{
	9:  helpGeneric9,
	10: helpGeneric10,
}

var cloudSpecificHelp = map[string]string{
	"aws-9":    helpAWS9,
	"aws-10":   helpAWS10,
	"azure-9":  helpAzure9,
	"azure-10": helpAzure10,
	"gce-9":    helpGCE9,
	"gce-10":   helpGCE10,
}

const fallbackHelp = `
Please ensure that this system is registered with Red Hat and has access to the required Red Hat repositories.
See: https://docs.redhat.com/en/documentation/subscription_central/1-latest/html/getting_started_with_rhel_system_registration/index
`

func getDetailedSubscriptionError(err error, systemInfo sysinfo.SysInfo) error {
	key := fmt.Sprintf("%s-%d", systemInfo.CloudProvider, systemInfo.OsVersion)
	message, ok := cloudSpecificHelp[key]
	if !ok {
		message, ok = genericHelp[systemInfo.OsVersion]
		if !ok {
			message = fallbackHelp
		}
	}
	return fmt.Errorf("failed to enable required repositories\n%s\nOriginal error: %w", message, err)
}
