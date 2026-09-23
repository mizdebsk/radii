package core

import "github.com/mizdebsk/radii/internal/api"

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
