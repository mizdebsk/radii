package main

import (
	"context"
	"os"

	"github.com/mizdebsk/radii/internal/api"
	"github.com/mizdebsk/radii/internal/cli"
	"github.com/mizdebsk/radii/internal/dnf"
	"github.com/mizdebsk/radii/internal/exec"
	"github.com/mizdebsk/radii/internal/provider/amd"
	"github.com/mizdebsk/radii/internal/provider/nvidia"
	"github.com/mizdebsk/radii/internal/rhsm"
	"github.com/mizdebsk/radii/internal/sysinfo"
)

// set at build time via -ldflags, eg: go build -ldflags="-X main.version=1.0.0" ./cmd/radii
var version = "dev"

func main() {
	ctx := context.Background()
	executor := exec.NewExecutor(ctx)
	systemInfo := sysinfo.DetectSysInfo()

	packageManager := dnf.NewPackageManager(executor)
	repositoryManager := rhsm.NewRepositoryManager(executor, systemInfo)
	providers := []api.Provider{nvidia.NewProvider(packageManager), amd.NewProvider(packageManager)}
	deps := api.CoreDeps{
		PackageManager:    packageManager,
		RepositoryManager: repositoryManager,
		Providers:         providers,
	}

	root := cli.NewRootCmd(deps, version)

	if err := root.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
