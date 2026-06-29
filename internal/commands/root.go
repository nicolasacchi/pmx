package commands

import (
	cliout "github.com/nicolasacchi/clicore/output"
	"github.com/nicolasacchi/pmx/internal/client"
	"github.com/nicolasacchi/pmx/internal/config"
	"github.com/nicolasacchi/pmx/internal/output"
	"github.com/nicolasacchi/pmx/internal/timeparse"
	"github.com/spf13/cobra"
)

var (
	version          = "dev"
	serverTokenFlag  string
	accountTokenFlag string
	projectFlag      string
	jsonFlag         bool
	jqFlag           string
	verboseFlag      bool
	quietFlag        bool
	fromFlag         string
	toFlag           string
	limitFlag        int
	offsetFlag       int
)

var rootCmd = &cobra.Command{
	Use:   "pmx",
	Short: "pmx — Postmark Explorer CLI",
	Long: `pmx is a CLI for the Postmark API — built for humans AND for LLM/agent use.

Humans get pretty tables on a TTY; pipes, --json, --jq, and Claude Code
(CLAUDECODE=1) get clean JSON with an agent row-cap + {"truncated":...} marker.

The headline use case — trace a passwordless OTP that a customer says never
arrived (OTP is sent over SMTP but Postmark tracks it the same way):
  pmx otp customer@example.com
  pmx messages outbound --recipient customer@example.com --from 2d
  pmx messages outbound events <message-id>

More examples:
  pmx bounces list --inactive
  pmx suppressions list --stream outbound
  pmx stats outbound --from 24h
  pmx overview
  pmx signatures list      # account scope (DKIM/SPF status)

Auth has two scopes (set via flags, env, or 'pmx config add'):
  server  → X-Postmark-Server-Token  (POSTMARK_SERVER_TOKEN) — sending, messages,
            bounces, stats, templates, streams, suppressions, server, webhooks
  account → X-Postmark-Account-Token (POSTMARK_ACCOUNT_TOKEN) — servers,
            signatures, domains, data-removals
Run 'pmx config doctor' to see which scopes your tokens can reach.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func SetVersion(v string) {
	version = v
	rootCmd.Version = v
}

func Execute() error {
	return rootCmd.Execute()
}

// getClient builds a Postmark client from the resolved credentials. Both tokens
// are optional; the per-scope check fires at request time with a clear hint.
func getClient(_ *cobra.Command) (*client.Client, error) {
	creds, err := config.LoadCredentials(serverTokenFlag, accountTokenFlag, projectFlag)
	if err != nil {
		return nil, err
	}
	return client.New(creds.ServerToken, creds.AccountToken, verboseFlag), nil
}

func isJSONMode() bool {
	return output.IsJSON(jsonFlag, jqFlag)
}

// printData applies the agent row-cap (shape-safe; explicit --limit wins) then
// renders a table (TTY + known command) or JSON.
func printData(command string, data []byte) error {
	data = cliout.CapAgentArray(data, limitFlag)
	return output.PrintData(command, data, isJSONMode(), jqFlag)
}

// fromDate / toDate resolve the --from / --to flags to Postmark YYYY-MM-DD
// filters (empty string when the flag is unset → no date filter).
func fromDate() (string, error) { return timeparse.ParseDate(fromFlag) }
func toDate() (string, error)   { return timeparse.ParseDate(toFlag) }

func init() {
	pf := rootCmd.PersistentFlags()
	pf.StringVar(&serverTokenFlag, "server-token", "", "Postmark Server token (overrides POSTMARK_SERVER_TOKEN)")
	pf.StringVar(&accountTokenFlag, "account-token", "", "Postmark Account token (overrides POSTMARK_ACCOUNT_TOKEN)")
	pf.StringVar(&projectFlag, "project", "", "Use a named project from the config file")
	pf.BoolVar(&jsonFlag, "json", false, "Force JSON output (auto-enabled when stdout is not a TTY)")
	pf.StringVar(&jqFlag, "jq", "", "Apply a gjson path filter to JSON output (gjson syntax, not jq)")
	pf.BoolVarP(&verboseFlag, "verbose", "v", false, "Print request/response details to stderr (PII redacted)")
	pf.BoolVarP(&quietFlag, "quiet", "q", false, "Suppress non-error output")
	pf.StringVar(&fromFlag, "from", "", "Date filter start (e.g. 7d, 24h, 2026-06-01, now); date-granular")
	pf.StringVar(&toFlag, "to", "", "Date filter end (e.g. now, 2026-06-29)")
	pf.IntVar(&limitFlag, "limit", 50, "Max results to return (Postmark 'count'; agent default 100 under CLAUDECODE)")
	pf.IntVar(&offsetFlag, "offset", 0, "Pagination offset (Postmark 'offset')")
}
