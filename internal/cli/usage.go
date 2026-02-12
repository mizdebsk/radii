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
  --verbose    Increase output verbosity
  --quiet      Suppress informational output
  --version    Show program version
`, progName, progName)
}

func printInstallUsage() {
	fmt.Printf(`Usage:
  %s install --auto-detect
  %s install <vendor[:version]>...

Install driver stacks.
Either specify drivers explicitly, or use --auto-detect to install
the default drivers for hardware detected in the system.

Options:
  --auto-detect   Select drivers automatically (exclusive with arguments)
  --batch         Run non-interactively
  --dry-run       Show what would be done without making changes
  --force         Install even if detection does not match hardware
`, progName, progName)
}

func printRemoveUsage() {
	fmt.Printf(`Usage:
  %s remove --all
  %s remove <vendor[:version]>...

Remove installed driver stacks.

Options:
  --all       Remove all managed drivers (exclusive with arguments)
  --batch     Run non-interactively
  --dry-run   Show what would be done without making changes
`, progName, progName)
}

func printListUsage() {
	fmt.Printf(`Usage:
  %s list [--available] [--installed]

List driver stacks.
If no options are given, --available is assumed.

Options:
  --available   Show drivers available from repositories (default)
  --installed   Show currently installed drivers
`, progName)
}
