package main

import (
	"context"
	"fmt"
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

	if err := cli.Execute(os.Args, deps, version); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
