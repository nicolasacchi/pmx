package commands

import (
	"net/url"

	"github.com/spf13/cobra"
)

// stats.go — outbound delivery statistics (server scope). Every subcommand is a
// read returning a single aggregate object (not a list), so all render via
// printData("", data). The parent owns a persistent --tag filter; --from/--to
// are global. Mirrors the messages.go group shape.

var statsTag string

func init() {
	rootCmd.AddCommand(statsCmd)
	statsCmd.PersistentFlags().StringVar(&statsTag, "tag", "", "Filter statistics by message tag (e.g. otp)")
	statsCmd.AddCommand(
		statsOutboundCmd,
		statsSendsCmd,
		statsBouncesCmd,
		statsSpamCmd,
		statsOpensCmd,
		statsClicksCmd,
	)
}

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Outbound delivery statistics (server scope)",
	Long: `Outbound delivery statistics for the server, aggregated over a date
range. Use --from/--to to bound the window and --tag to scope to one message tag.

Examples:
  pmx stats outbound --from 24h
  pmx stats opens --from 7d --tag otp`,
}

// statsRun builds the shared param set (date range + tag), GETs the given path,
// and prints the aggregate object.
func statsRun(path string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		params := url.Values{}
		if err := addDateParams(params); err != nil {
			return err
		}
		setIfNonEmpty(params, "tag", statsTag)

		data, err := serverGet(c, path, params)
		if err != nil {
			return err
		}
		return printData("", data)
	}
}

var statsOutboundCmd = &cobra.Command{
	Use:   "outbound",
	Short: "Aggregate sent/bounce/spam/open totals",
	RunE:  statsRun("stats/outbound"),
}

var statsSendsCmd = &cobra.Command{
	Use:   "sends",
	Short: "Counts of sent messages over the date range",
	RunE:  statsRun("stats/outbound/sends"),
}

var statsBouncesCmd = &cobra.Command{
	Use:   "bounces",
	Short: "Counts of bounced messages over the date range",
	RunE:  statsRun("stats/outbound/bounces"),
}

var statsSpamCmd = &cobra.Command{
	Use:   "spam",
	Short: "Counts of spam complaints over the date range",
	RunE:  statsRun("stats/outbound/spam"),
}

var statsOpensCmd = &cobra.Command{
	Use:   "opens",
	Short: "Counts of message opens over the date range",
	RunE:  statsRun("stats/outbound/opens"),
}

var statsClicksCmd = &cobra.Command{
	Use:   "clicks",
	Short: "Counts of link clicks over the date range",
	RunE:  statsRun("stats/outbound/clicks"),
}
