package api

//go:generate mockgen -source=provider.go -destination=../mocks/provider_mock.go -package=mocks

type Provider interface {
	GetID() string
	GetName() string
	RequiredChannels() []string
	Install(drivers []DriverID, kernel KernelTarget) ([]string, error)
	// An empty version selects all installed versions, including leftover components.
	Remove(drivers []DriverID) ([]string, error)
	// The first available driver is the provider's default.
	ListAvailable() ([]DriverID, error)
	ListInstalled() ([]DriverID, error)
	DetectHardware() (bool, error)
}
