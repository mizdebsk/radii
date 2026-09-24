package core

import (
	"fmt"
	"strings"

	"github.com/mizdebsk/radii/internal/api"
	"github.com/mizdebsk/radii/internal/log"
)

func RemoveSpecific(deps api.CoreDeps, drivers []string, batchMode, dryRun bool) error {
	var toRemove []api.DriverID

	if len(drivers) == 0 {
		return fmt.Errorf("not specified what to remove")
	}
outer:
	for _, driverStr := range drivers {
		allVersions := driverStr != "" && !strings.Contains(driverStr, ":")
		driver := api.DriverID{ProviderID: driverStr}
		if !allVersions {
			var err error
			driver, err = parseDriverID(driverStr)
			if err != nil {
				return err
			}
		}
		provider, err := lookupProvider(deps, driver)
		if err != nil {
			return err
		}
		if allVersions {
			toRemove = append(toRemove, driver)
			continue
		}
		installed, err := provider.ListInstalled()
		if err != nil {
			return fmt.Errorf("failed to list installed %s drivers: %w", provider.GetName(), err)
		}
		for _, inst := range installed {
			if inst.Version == driver.Version {
				toRemove = append(toRemove, inst)
				continue outer
			}
		}
		return fmt.Errorf("driver %s version %s is NOT installed", provider.GetName(), driver.Version)
	}
	return doRemove(deps, toRemove, batchMode, dryRun)
}

func RemoveAll(deps api.CoreDeps, batchMode, dryRun bool) error {
	var toRemove []api.DriverID

	for _, provider := range deps.Providers {
		toRemove = append(toRemove, api.DriverID{ProviderID: provider.GetID()})
	}
	return doRemove(deps, toRemove, batchMode, dryRun)
}

func doRemove(deps api.CoreDeps, toRemove []api.DriverID, batchMode, dryRun bool) error {
	var allPkgs []string
	for _, provider := range deps.Providers {
		provID := provider.GetID()
		var provToRemove []api.DriverID
		for _, driver := range toRemove {
			if driver.ProviderID == provID {
				provToRemove = append(provToRemove, driver)
			}
		}
		if len(provToRemove) != 0 {
			pkgs, err := provider.Remove(provToRemove)
			if err != nil {
				return fmt.Errorf("failed to remove %s driver: %w", provider.GetName(), err)
			}
			allPkgs = append(allPkgs, pkgs...)
		}
	}
	if len(allPkgs) == 0 {
		return fmt.Errorf("nothing to remove")
	}
	for _, pkg := range allPkgs {
		log.Logf("package will be removed: %v", pkg)
	}
	if err := deps.PackageManager.Remove(allPkgs, batchMode, dryRun); err != nil {
		return fmt.Errorf("failed to remove pacakges: %w", err)
	}
	return nil
}
