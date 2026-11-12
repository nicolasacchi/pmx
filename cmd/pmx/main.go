// Command pmx is the Postmark Explorer CLI — a fleet sibling of ddx/otx/gx/stx.
// It gives humans pretty tables and LLMs/Claude Code clean JSON over the full
// Postmark API: message search & delivery tracing (the OTP lens), bounces,
// suppressions, stats, templates, plus account-level signature/domain admin.
package main

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/nicolasacchi/pmx/internal/client"
	"github.com/nicolasacchi/pmx/internal/commands"
	"github.com/nicolasacchi/pmx/internal/output"
)

// version is overwritten at build time via -ldflags "-X main.version=...".
var version = "dev"

// resolveVersion returns the ldflags-injected version, or, for `go install
// ...@<tag>` builds where no ldflags are set, the module version recorded in
// the binary's build info.
func resolveVersion() string {
	if version != "dev" {
		return version
	}
	if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		return bi.Main.Version
	}
	return version
}

func main() {
	commands.SetVersion(resolveVersion())
	if err := commands.Execute(); err != nil {
		var apiErr *client.APIError
		if errors.As(err, &apiErr) {
			output.PrintError(apiErr.Error(), apiErr.StatusCode)
			fmt.Fprintln(os.Stderr, "Error:", apiErr.Error())
			if apiErr.Hint != "" {
				fmt.Fprintln(os.Stderr, "Hint:", apiErr.Hint)
			}
			os.Exit(apiErr.ExitCode())
		}
		output.PrintError(err.Error(), 0)
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
