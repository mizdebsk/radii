package cli

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/mizdebsk/radii/internal/api"
	"github.com/mizdebsk/radii/internal/core"
	"github.com/mizdebsk/radii/internal/log"
)

var progName = ""

func Execute(argv []string, deps api.CoreDeps, version string) error {
	progName = filepath.Base(argv[0])
	args := argv[1:]
	if helpRequested(args) {
		printRootUsage()
		return nil
	}

	var showVersion bool

	fs := flag.NewFlagSet(progName, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {}
	fs.BoolVar(&log.Verbose, "verbose", false, "")
	fs.BoolVar(&log.Quiet, "quiet", false, "")
	fs.BoolVar(&log.Debug, "debug", false, "")
	fs.BoolVar(&showVersion, "version", false, "")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if showVersion {
		printVersion(progName, version)
		return nil
	}

	rest := fs.Args()
	if len(rest) == 0 {
		return fmt.Errorf("no command specified")
	}

	switch rest[0] {
	case "install", "in":
		return runInstall(rest[1:], deps)
	case "remove", "rm":
		return runRemove(rest[1:], deps)
	case "list", "ls":
		return runList(rest[1:], deps)
	default:
		return fmt.Errorf("unknown command: %s", rest[0])
	}
}

func runInstall(args []string, deps api.CoreDeps) error {
	if helpRequested(args) {
		printInstallUsage()
		return nil
	}

	var (
		autoDetect bool
		batchMode  bool
		dryRun     bool
		force      bool
	)

	fs := flag.NewFlagSet("install", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.BoolVar(&autoDetect, "auto-detect", false, "")
	fs.BoolVar(&batchMode, "batch", false, "")
	fs.BoolVar(&dryRun, "dry-run", false, "")
	fs.BoolVar(&force, "force", false, "")
	if err := fs.Parse(args); err != nil {
		return err
	}

	drivers := fs.Args()

	if autoDetect {
		if len(drivers) > 0 {
			return fmt.Errorf("both --auto-detect and specific drivers given")
		}
		if force {
			return fmt.Errorf("both --auto-detect and --force were specified")
		}
		return core.InstallAutoDetect(deps, batchMode, dryRun)
	}

	if len(drivers) == 0 {
		return fmt.Errorf("not specified what to install (use --auto-detect or provide drivers)")
	}

	return core.InstallSpecific(deps, drivers, batchMode, dryRun, force)
}

func runRemove(args []string, deps api.CoreDeps) error {
	if helpRequested(args) {
		printRemoveUsage()
		return nil
	}

	var (
		all       bool
		batchMode bool
		dryRun    bool
	)

	fs := flag.NewFlagSet("remove", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.BoolVar(&all, "all", false, "")
	fs.BoolVar(&batchMode, "batch", false, "")
	fs.BoolVar(&dryRun, "dry-run", false, "")
	if err := fs.Parse(args); err != nil {
		return err
	}

	drivers := fs.Args()

	if all {
		if len(drivers) > 0 {
			return fmt.Errorf("both --all and specific drivers given")
		}
		return core.RemoveAll(deps, batchMode, dryRun)
	}

	if len(drivers) == 0 {
		return fmt.Errorf("not specified what to remove (use --all or provide drivers)")
	}

	return core.RemoveSpecific(deps, drivers, batchMode, dryRun)
}

func runList(args []string, deps api.CoreDeps) error {
	if helpRequested(args) {
		printListUsage()
		return nil
	}

	var (
		flagAvailable bool
		flagInstalled bool
	)

	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.BoolVar(&flagAvailable, "available", false, "")
	fs.BoolVar(&flagInstalled, "installed", false, "")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if len(fs.Args()) != 0 {
		return fmt.Errorf("list takes no arguments")
	}

	if flagAvailable || (!flagAvailable && !flagInstalled) {
		res, err := core.List(deps, true, true, true)
		if err != nil {
			return err
		}

		if len(res) > 0 {
			fmt.Println("Available drivers:")
			for _, dev := range res {
				markInstalled := " "
				if dev.Installed {
					markInstalled = "*"
				}
				markAuto := " "
				if dev.Compatible {
					markAuto = ">"
				}
				fmt.Printf("%s%s %s:%s\n", markInstalled, markAuto, dev.ID.ProviderID, dev.ID.Version)
			}
		} else {
			fmt.Println("Available drivers:\n  (none)")
		}
	}

	if flagInstalled {
		res, err := core.List(deps, true, false, false)
		if err != nil {
			return err
		}

		fmt.Print("Installed drivers:")
		for _, dev := range res {
			if dev.Installed {
				fmt.Printf("\n%s:%s", dev.ID.ProviderID, dev.ID.Version)
			}
		}
		fmt.Println()
	}

	return nil
}

func helpRequested(args []string) bool {
	for _, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			return false
		}
		if arg == "--help" {
			return true
		}
	}
	return false
}

func printVersion(progName, version string) {
	v := strings.TrimSpace(version)
	if v == "" {
		v = "unknown"
	}
	fmt.Println(progName, "version", v)
}
