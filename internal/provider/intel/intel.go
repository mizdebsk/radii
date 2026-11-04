package intel

import (
	"fmt"

	"github.com/mizdebsk/radii/internal/api"
	"github.com/mizdebsk/radii/internal/hwdetect"
)

const (
	variantLatest = "latest"
	pkgOpenVINO   = "openvino"
)

type prov struct {
	PM api.PackageManager
}

var _ api.Provider = (*prov)(nil)

func NewProvider(pm api.PackageManager) api.Provider {
	return &prov{PM: pm}
}

func (p *prov) GetID() string {
	return "openvino"
}

func (p *prov) GetName() string {
	return "Intel GPU or accelerator"
}

func (p *prov) RequiredChannels() []string {
	return []string{"BaseOS", "AppStream", "Extensions"}
}

func (p *prov) Install(drivers []api.DriverID, kernel api.KernelTarget) ([]string, error) {
	if len(drivers) == 0 {
		return []string{}, nil
	}
	if kernel.Version != "" {
		return nil, fmt.Errorf("OpenVINO does not support selecting a specific kernel version")
	}
	for _, driver := range drivers {
		if driver.Version != variantLatest {
			return nil, fmt.Errorf("unknown OpenVINO variant: %s", driver.Version)
		}
	}
	return []string{pkgOpenVINO}, nil
}

func (p *prov) ListAvailable() ([]api.DriverID, error) {
	return p.driversForPackages(p.PM.ListAvailablePackages())
}

func (p *prov) ListInstalled() ([]api.DriverID, error) {
	return p.driversForPackages(p.PM.ListInstalledPackages())
}

func (p *prov) driversForPackages(packages []api.PackageInfo, err error) ([]api.DriverID, error) {
	if err != nil {
		return nil, err
	}
	for _, pkg := range packages {
		if pkg.Name == pkgOpenVINO {
			return []api.DriverID{{ProviderID: p.GetID(), Version: variantLatest}}, nil
		}
	}
	return []api.DriverID{}, nil
}

func (p *prov) Remove(drivers []api.DriverID) ([]string, error) {
	if len(drivers) == 0 {
		return []string{}, nil
	}
	for _, driver := range drivers {
		if driver.Version != "" && driver.Version != variantLatest {
			return nil, fmt.Errorf("unknown OpenVINO variant: %s", driver.Version)
		}
	}
	installed, err := p.ListInstalled()
	if err != nil {
		return nil, fmt.Errorf("failed to list installed packages: %w", err)
	}
	if len(installed) == 0 {
		return []string{}, nil
	}
	return []string{pkgOpenVINO}, nil
}

func (p *prov) DetectHardware() (bool, error) {
	return detectHardware(hwdetect.DefaultModaliasRoot)
}
