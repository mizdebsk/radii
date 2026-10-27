package api

import "github.com/mizdebsk/radii/internal/sysinfo"

//go:generate mockgen -source=core.go -destination=../mocks/core_mock.go -package=mocks

type RepositoryManager interface {
	SetSubscriptionsEnabled(enabled bool)
	SubscriptionsAvailable() bool
	EnsurePartnerReposEnabled(providerID string, repos []PartnerRepo) error
	PartnerRepoDNFArgs(repos []PartnerRepo) ([]string, error)
	EnsureRepositoriesEnabled(channels []string) error
	GetRepoIDs(channels []string) ([]string, error)
}

type PartnerRepo struct {
	RepoID  string
	Name    string
	BaseURL string
	GPGKey  string
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
