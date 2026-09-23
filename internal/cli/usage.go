package cli

import "fmt"

func printRootUsage() {
	fmt.Printf(`Usage:
  %s <command> [options]

Commands:
  install    Install driver stacks
  remove     Remove driver stacks
  list       List driver stacks

Run '%s <command> --help' for details.

Options:
  --verbose              Increase output verbosity
  --quiet                Suppress informational output
  --skip-subscriptions   Skip Red Hat Subscription Manager (RHSM) setup
  --version              Show program version
`, progName, progName)
}

func printInstallUsage() {
	fmt.Printf(`Usage:
  %s install [options] --auto-detect
  %s install [options] <vendor[:version]>...

Install driver stacks.
Omit the version or use :default to select the provider's default driver stack.
Use :latest to select the newest available version.
Either specify drivers explicitly, or use --auto-detect to install
the default drivers for hardware detected in the system.
Installing drivers from multiple providers may cause conflicts.
For multiple providers, choose drivers explicitly or add --force to install all.

Options:
  --auto-detect              Select drivers automatically (exclusive with arguments)
  --batch                    Run non-interactively
  --dry-run                  Preview without changing packages or repository configuration
  --force                    Install even if detection does not match hardware
                             With --auto-detect, install drivers for all detected providers
  -K, --kernel RELEASE       Target a specific kernel version-release[.arch][+variant]
  --kernel-variant VARIANT   Select default (also 4k or empty) or 64k

NVIDIA defaults to the running kernel variant and the latest matching kernel.
A variant embedded in --kernel overrides the running variant.
AMD ignores the variant and does not support --kernel.
`, progName, progName)
}

func printRemoveUsage() {
	fmt.Printf(`Usage:
  %s remove --all
  %s remove <vendor[:version]>...

Remove installed driver stacks.
Omit the version to remove all installed versions for that provider.

Options:
  --all       Remove all managed drivers (exclusive with arguments)
  --batch     Run non-interactively
  --dry-run   Show what would be done without making changes
`, progName, progName)
}

func printListUsage() {
	fmt.Printf(`Usage:
  %s list [--available] [--installed] [--compatible]

List driver stacks.
If no options are given, --available is assumed.

Options:
  --available   Show drivers available from repositories (default)
  --installed   Show currently installed drivers
  --compatible  Only list compatible drivers
`, progName)
}
