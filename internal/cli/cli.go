package cli

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/mizdebsk/radii/internal/api"
	"github.com/mizdebsk/radii/internal/config"
	"github.com/mizdebsk/radii/internal/core"
	"github.com/mizdebsk/radii/internal/log"
)

var progName = ""

func enablePartnerRepos(deps api.CoreDeps) {
	for _, p := range deps.Providers {
		type partnerSetter interface {
			SetPartnerRepoEnabled(bool)
		}
		if ps, ok := p.(partnerSetter); ok {
			ps.SetPartnerRepoEnabled(true)
		}
	}
}

func Execute(argv []string, deps api.CoreDeps, version string) error {
	progName = filepath.Base(argv[0])
	args := argv[1:]
	if helpRequested(args) {
		printRootUsage()
		return nil
	}

	var (
		showVersion       bool
		skipSubscriptions bool
		partnerRepo       bool
	)

	fs := flag.NewFlagSet(progName, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {}
	fs.BoolVar(&log.Verbose, "verbose", false, "")
	fs.BoolVar(&log.Quiet, "quiet", false, "")
	fs.BoolVar(&log.Debug, "debug", false, "")
	fs.BoolVar(&showVersion, "version", false, "")
	fs.BoolVar(&skipSubscriptions, "skip-subscriptions", false, "")
	fs.BoolVar(&partnerRepo, "partner-repo", false, "")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if showVersion {
		printVersion(progName, version)
		return nil
	}

	cfg, err := config.Load(config.DefaultConfigPath)
	if err != nil {
		return err
	}

	if skipSubscriptions {
		deps.RepositoryManager.SetSubscriptionsEnabled(false)
	}

	partnerRepoExplicit := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "partner-repo" {
			partnerRepoExplicit = true
		}
	})
	usePartnerRepo := cfg.PartnerRepos
	if partnerRepoExplicit {
		usePartnerRepo = partnerRepo
	}
	if usePartnerRepo {
		enablePartnerRepos(deps)
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
	case "config":
		return runConfig(rest[1:])
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
		kernel     api.KernelOptions
	)

	fs := flag.NewFlagSet("install", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.BoolVar(&autoDetect, "auto-detect", false, "")
	fs.BoolVar(&batchMode, "batch", false, "")
	fs.BoolVar(&dryRun, "dry-run", false, "")
	fs.BoolVar(&force, "force", false, "")
	fs.StringVar(&kernel.Version, "kernel", "", "")
	fs.StringVar(&kernel.Version, "K", "", "")
	fs.Func("kernel-variant", "", func(value string) error {
		kernel.Variant = &value
		return nil
	})
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			printInstallUsage()
			return nil
		}
		return err
	}

	drivers := fs.Args()

	if autoDetect {
		if len(drivers) > 0 {
			return fmt.Errorf("both --auto-detect and specific drivers given")
		}
		return core.InstallAutoDetect(deps, batchMode, dryRun, force, kernel)
	}

	if len(drivers) == 0 {
		return fmt.Errorf("not specified what to install (use --auto-detect or provide drivers)")
	}

	return core.InstallSpecific(deps, drivers, batchMode, dryRun, force, kernel)
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
		flagAvailable  bool
		flagInstalled  bool
		flagCompatible bool
	)

	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.BoolVar(&flagAvailable, "available", false, "")
	fs.BoolVar(&flagInstalled, "installed", false, "")
	fs.BoolVar(&flagCompatible, "compatible", false, "")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if len(fs.Args()) != 0 {
		return fmt.Errorf("list takes no arguments")
	}

	if flagAvailable || (!flagAvailable && !flagInstalled) {
		res, err := core.List(deps, true, true, true, flagCompatible)
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
		res, err := core.List(deps, true, false, flagCompatible, flagCompatible)
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

func runConfig(args []string) error {
	if helpRequested(args) {
		printConfigUsage()
		return nil
	}
	if len(args) == 0 {
		cfg, err := config.Load(config.DefaultConfigPath)
		if err != nil {
			return err
		}
		fmt.Printf("partner-repos: %t\n", cfg.PartnerRepos)
		return nil
	}
	if len(args) == 1 {
		cfg, err := config.Load(config.DefaultConfigPath)
		if err != nil {
			return err
		}
		switch args[0] {
		case "partner-repos":
			fmt.Printf("%t\n", cfg.PartnerRepos)
		default:
			return fmt.Errorf("unknown config key: %s", args[0])
		}
		return nil
	}
	if len(args) != 2 {
		return fmt.Errorf("usage: %s config <key> [<value>]", progName)
	}
	key, value := args[0], args[1]
	cfg, err := config.Load(config.DefaultConfigPath)
	if err != nil {
		return err
	}
	switch key {
	case "partner-repos":
		v := strings.ToLower(value)
		switch v {
		case "true", "1", "yes":
			cfg.PartnerRepos = true
		case "false", "0", "no":
			cfg.PartnerRepos = false
		default:
			return fmt.Errorf("invalid value for partner-repos: %q (expected true/false)", value)
		}
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}
	return config.Save(config.DefaultConfigPath, cfg)
}

func printVersion(progName, version string) {
	v := strings.TrimSpace(version)
	if v == "" {
		v = "unknown"
	}
	fmt.Println(progName, "version", v)
}
