package dnf

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/mizdebsk/rhel-drivers/internal/api"
	"github.com/mizdebsk/rhel-drivers/internal/cache"
	"github.com/mizdebsk/rhel-drivers/internal/log"
)

const defaultDNFBinary = "dnf"

type pkgMgr struct {
	Bin string
}

var _ api.PackageManager = (*pkgMgr)(nil)

func New() api.PackageManager {
	return &pkgMgr{
		Bin: defaultDNFBinary,
	}
}

var availableCache = cache.Cache[[]api.PackageInfo]{}
var installedCache = cache.Cache[[]api.PackageInfo]{}

func (pm *pkgMgr) ListAvailablePackages(ctx context.Context) ([]api.PackageInfo, error) {
	return availableCache.Get(ctx, func(ctx context.Context) ([]api.PackageInfo, error) {
		tags := []string{"name", "epoch", "version", "release", "arch", "sourcerpm", "repoid"}
		// QQQ and YYY are there to make filtering spurious lines easier.
		format := "QQQ"
		for _, field := range tags {
			format += "|%{" + field + "}"
		}
		// Trailing NL is not required with DNF 4, but will be required with DNF 5.
		// With DNF 4 it will result in empty lines, but they are ignored anyway.
		format += "|YYY\n"
		return runQueryCommand(ctx, pm.Bin, "-q", "repoquery", "--qf", format)
	})
}

func (pm *pkgMgr) ListInstalledPackages(ctx context.Context) ([]api.PackageInfo, error) {
	return installedCache.Get(ctx, func(ctx context.Context) ([]api.PackageInfo, error) {
		tags := []string{"NAME", "EPOCH", "VERSION", "RELEASE", "ARCH", "SOURCERPM"}
		format := "QQQ"
		for _, field := range tags {
			format += "|%|" + field + "?{%{" + field + "}}|"
		}
		format += "||YYY\n"
		return runQueryCommand(ctx, "rpm", "-qa", "--qf", format)
	})
}

func runQueryCommand(ctx context.Context, command string, args ...string) ([]api.PackageInfo, error) {
	cmd := exec.CommandContext(ctx, command, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout for %s command: %w", command, err)
	}
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start %s command: %w", command, err)
	}

	var infos []api.PackageInfo
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "QQQ|") && strings.HasSuffix(line, "|YYY") {
			line = strings.TrimPrefix(line, "QQQ|")
			line = strings.TrimSuffix(line, "|YYY")
			fields := strings.SplitN(line, "|", 7)
			if len(fields) == 7 {
				infos = append(infos, api.PackageInfo{
					Name:       fields[0],
					Epoch:      fields[1],
					Version:    fields[2],
					Release:    fields[3],
					Arch:       fields[4],
					SourceName: parseSourceName(fields[5]),
					Repo:       fields[6],
				})
			}
		}
	}

	if err := scanner.Err(); err != nil {
		_ = cmd.Wait()
		return nil, fmt.Errorf("error reading %s output: %w", command, err)
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("%s command failed: %w", command, err)
	}

	return infos, nil
}

func parseSourceName(sourceRpm string) string {
	sourceRpm = strings.TrimSpace(sourceRpm)
	if sourceRpm == "" {
		return ""
	}
	s := strings.TrimSuffix(sourceRpm, ".rpm")
	parts := strings.Split(s, "-")
	if len(parts) < 2 {
		return s
	}
	if len(parts) >= 3 {
		return strings.Join(parts[:len(parts)-2], "-")
	}
	return parts[0]
}

func (pm *pkgMgr) Install(ctx context.Context, packages []string, opts api.InstallOptions) error {
	if len(packages) != 0 {
		log.Logf("installing packages: %v", packages)
		if !opts.DryRun {
			return runDnfCommand(ctx, pm.Bin, append([]string{"install"}, packages...))
		}
	} else {
		log.Warnf("no packages to install")
	}
	return nil
}

func (pm *pkgMgr) Remove(ctx context.Context, packages []string, opts api.RemoveOptions) error {
	if len(packages) != 0 {
		log.Logf("removing packages: %v", packages)
		if !opts.DryRun {
			return runDnfCommand(ctx, pm.Bin, append([]string{"remove"}, packages...))
		}
	} else {
		log.Warnf("no packages to remove")
	}
	return nil
}

func runDnfCommand(ctx context.Context, command string, args []string) error {
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s command failed: %w", command, err)
	}
	return nil
}
