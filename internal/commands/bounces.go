package commands

import (
	"net/url"

	"github.com/nicolasacchi/pmx/internal/client"
	"github.com/spf13/cobra"
)

var (
	bouncesType      string
	bouncesEmail     string
	bouncesTag       string
	bouncesMessageID string
	bouncesInactive  bool
)

func init() {
	rootCmd.AddCommand(bouncesCmd)
	bouncesCmd.AddCommand(
		bouncesListCmd,
		bouncesGetCmd,
		bouncesDumpCmd,
		bouncesActivateCmd,
		bouncesTagsCmd,
		bouncesDeliveryStatsCmd,
	)

	lf := bouncesListCmd.Flags()
	lf.StringVar(&bouncesType, "type", "", "Filter by bounce type (e.g. HardBounce, SpamComplaint, Transient)")
	lf.StringVar(&bouncesEmail, "email", "", "Filter by recipient email")
	lf.StringVar(&bouncesTag, "tag", "", "Filter by message tag")
	lf.StringVar(&bouncesMessageID, "message-id", "", "Filter by originating message id")
	lf.BoolVar(&bouncesInactive, "inactive", false, "Only show bounces that deactivated the recipient")
}

var bouncesCmd = &cobra.Command{
	Use:   "bounces",
	Short: "Inspect and reactivate bounces (server scope)",
}

var bouncesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List bounces with optional type/email/tag filters",
	Long: `List bounces for the server. A bounce that deactivates a recipient stops
future sends to that address until reactivated (see 'bounces activate').

Examples:
  pmx bounces list --inactive
  pmx bounces list --type HardBounce --from 7d
  pmx bounces list --email customer@example.com
  pmx bounces list --tag otp --jq '#.{id:ID,type:Type,email:Email,inactive:Inactive}'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		params := pageParams()
		if err := addDateParams(params); err != nil {
			return err
		}
		setIfNonEmpty(params, "type", bouncesType)
		setIfNonEmpty(params, "emailFilter", bouncesEmail)
		setIfNonEmpty(params, "tag", bouncesTag)
		setIfNonEmpty(params, "messageID", bouncesMessageID)
		if bouncesInactive {
			params.Set("inactive", "true")
		}

		data, err := serverGet(c, "bounces", params)
		if err != nil {
			return err
		}
		return printData("bounces.list", extractList(data, "Bounces"))
	},
}

var bouncesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get full details for one bounce",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		data, err := serverGet(c, "bounces/"+url.PathEscape(args[0]), nil)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}

var bouncesDumpCmd = &cobra.Command{
	Use:   "dump <id>",
	Short: "Get the raw SMTP dump for a bounce",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		data, err := serverGet(c, "bounces/"+url.PathEscape(args[0])+"/dump", nil)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}

var bouncesActivateCmd = &cobra.Command{
	Use:   "activate <id>",
	Short: "Reactivate a deactivated bounce so sending to the recipient resumes",
	Long: `Reactivate a bounce, clearing the deactivation that blocks further sends to
the recipient. This is a write operation and requires --yes (or --confirm).

Examples:
  pmx bounces activate 123456789 --yes
  pmx bounces activate 123456789 --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if handled, err := gateWrite("activate bounce " + args[0]); handled {
			return err
		}
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		data, err := c.Put(ctx(), client.ScopeServer, "bounces/"+url.PathEscape(args[0])+"/activate", nil)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}

var bouncesTagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "List tags that have produced bounces",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		data, err := serverGet(c, "bounces/tags", nil)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}

var bouncesDeliveryStatsCmd = &cobra.Command{
	Use:   "deliverystats",
	Short: "Show delivery stats and bounce-type breakdown for the server",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		data, err := serverGet(c, "deliverystats", nil)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}
