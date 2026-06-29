package commands

import (
	"fmt"
	"os"

	"github.com/nicolasacchi/pmx/internal/client"
	"golang.org/x/term"
)

// Write-safety flags. State-changing Postmark verbs (bounce activate,
// suppression create/delete, *delete, email send, data-removal create) refuse
// unless the operator opts in. Registered as persistent flags on rootCmd.
var (
	yesFlag    bool
	dryRunFlag bool
)

func init() {
	rootCmd.PersistentFlags().BoolVar(&yesFlag, "yes", false, "Confirm destructive operations (alias: --confirm)")
	rootCmd.PersistentFlags().BoolVar(&yesFlag, "confirm", false, "Alias for --yes")
	rootCmd.PersistentFlags().BoolVar(&dryRunFlag, "dry-run", false, "Print the intended mutation and exit without sending")
}

// requireConfirm gates a state-changing verb, returning a write_locked APIError
// (exit 6) when --yes/--confirm was not passed, so agents can dispatch on Kind.
func requireConfirm(action string) error {
	if yesFlag {
		return nil
	}
	hint := "re-run with --yes (or --confirm) to proceed"
	if term.IsTerminal(int(os.Stdout.Fd())) {
		hint = "re-run with --yes to proceed, or --dry-run to preview"
	}
	return &client.APIError{
		Kind:   "write_locked",
		Detail: fmt.Sprintf("%s requires confirmation", action),
		Hint:   hint,
	}
}

// dryRun reports whether --dry-run was passed; gated verbs should print their
// intended mutation and return nil (exit 0) without sending a request.
func dryRun() bool { return dryRunFlag }
