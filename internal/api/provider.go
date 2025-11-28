package api

type Provider interface {
	GetID() string
	GetName() string
	Install(drivers []DriverID) ([]string, error)
	Remove(drivers []DriverID) ([]string, error)
	ListAvailable() ([]DriverID, error)
	ListInstalled() ([]DriverID, error)
	DetectHardware() (bool, error)
}
