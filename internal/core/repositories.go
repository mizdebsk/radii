package core

import (
	"fmt"

	"github.com/mizdebsk/radii/internal/api"
)

func prepareRepositories(deps api.CoreDeps, providers []api.Provider, readOnly bool) error {
	channels := requiredChannels(providers)
	partnerRepos := requiredPartnerRepos(providers)

	subsAvailable := deps.RepositoryManager.SubscriptionsAvailable()
	if len(partnerRepos) > 0 && !subsAvailable {
		partnerRepos = nil
	}

	if !readOnly {
		for _, provider := range providers {
			provRepos := provider.RequiredPartnerRepos()
			if len(provRepos) > 0 && subsAvailable {
				if err := deps.RepositoryManager.EnsurePartnerReposEnabled(provider.GetID(), provRepos); err != nil {
					return fmt.Errorf("failed to configure partner repositories: %w", err)
				}
			}
		}
		if err := deps.RepositoryManager.EnsureRepositoriesEnabled(channels); err != nil {
			return fmt.Errorf("failed to verify/enable repositories: %w", err)
		}
	}

	repos, err := deps.RepositoryManager.GetRepoIDs(channels)
	if err != nil {
		return fmt.Errorf("failed to determine query repositories: %w", err)
	}
	for _, pr := range partnerRepos {
		repos = append(repos, pr.RepoID)
	}
	deps.PackageManager.SetEnableRepos(repos)

	if readOnly && len(partnerRepos) > 0 {
		dnfArgs, err := deps.RepositoryManager.PartnerRepoDNFArgs(partnerRepos)
		if err != nil {
			return fmt.Errorf("failed to build partner repository arguments: %w", err)
		}
		deps.PackageManager.SetRepoArgs(dnfArgs)
	}

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

func requiredPartnerRepos(providers []api.Provider) []api.PartnerRepo {
	var repos []api.PartnerRepo
	for _, provider := range providers {
		repos = append(repos, provider.RequiredPartnerRepos()...)
	}
	return repos
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
