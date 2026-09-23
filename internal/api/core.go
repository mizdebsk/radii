package api

import "github.com/mizdebsk/radii/internal/sysinfo"

//go:generate mockgen -source=core.go -destination=../mocks/core_mock.go -package=mocks

type RepositoryManager interface {
	SetSubscriptionsEnabled(enabled bool)
	EnsureRepositoriesEnabled(channels []string) error
	GetRepoIDs(channels []string) ([]string, error)
}

type DriverID struct {
	ProviderID string
	Version    string
}

type CoreDeps struct {
	SystemInfo        sysinfo.SysInfo
	PackageManager    PackageManager
	RepositoryManager RepositoryManager
	Providers         []Provider
	Executor          Executor
}

type KernelOptions struct {
	Version string
	Variant *string
}

type KernelTarget struct {
	Version string
	Variant string
	Arch    string
}

type DriverStatus struct {
	ID         DriverID
	Available  bool
	Installed  bool
	Compatible bool
}
