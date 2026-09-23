package core

import (
	"fmt"

	"github.com/mizdebsk/radii/internal/api"
)

func prepareRepositories(deps api.CoreDeps, providers []api.Provider, readOnly bool) error {
	channels := requiredChannels(providers)
	var repos []string
	if readOnly {
		var err error
		repos, err = deps.RepositoryManager.GetRepoIDs(channels)
		if err != nil {
			return fmt.Errorf("failed to determine query repositories: %w", err)
		}
	} else if err := deps.RepositoryManager.EnsureRepositoriesEnabled(channels); err != nil {
		return fmt.Errorf("failed to verify/enable repositories: %w", err)
	}
	deps.PackageManager.SetEnableRepos(repos)
	return nil
}

func requiredChannels(providers []api.Provider) []string {
	var channels []string
	seen := make(map[string]bool)
	for _, provider := range providers {
		for _, channel := range provider.RequiredChannels() {
			if !seen[channel] {
				channels = append(channels, channel)
				seen[channel] = true
			}
		}
	}
	return channels
}

func providersForDrivers(providers []api.Provider, drivers []api.DriverID) []api.Provider {
	var selected []api.Provider
	for _, provider := range providers {
		for _, driver := range drivers {
			if provider.GetID() == driver.ProviderID {
				selected = append(selected, provider)
				break
			}
		}
	}
	return selected
}
