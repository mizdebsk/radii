package api

type RepositoryManager interface {
	EnsureRepositoriesEnabled() error
}

type DriverID struct {
	ProviderID string
	Version    string
}

type CoreDeps struct {
	PackageManager    PackageManager
	RepositoryManager RepositoryManager
	Providers         []Provider
	Executor          Executor
}

type DriverStatus struct {
	ID         DriverID
	Available  bool
	Installed  bool
	Compatible bool
}
