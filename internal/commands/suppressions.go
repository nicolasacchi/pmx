package commands

import (
	"github.com/nicolasacchi/pmx/internal/client"
	"github.com/spf13/cobra"
)

// suppressions.go follows the messages.go exemplar shape: one file per group,
// package-prefixed identifiers, registered in init() via rootCmd.AddCommand.
//
// Suppressions are managed per message stream (server scope). The group's
// --stream flag (default "outbound") selects which stream's suppression list to
// read or mutate. Suppressing a recipient stops Postmark from sending to them on
// that stream; deleting a suppression reactivates them.

var suppressionsStream string

var (
	suppressionsCreateEmail string
	suppressionsDeleteEmail string
)

func init() {
	rootCmd.AddCommand(suppressionsCmd)
	suppressionsCmd.AddCommand(suppressionsListCmd, suppressionsCreateCmd, suppressionsDeleteCmd)

	suppressionsCmd.PersistentFlags().StringVar(&suppressionsStream, "stream", "outbound", "Message stream id to manage suppressions for")

	suppressionsCreateCmd.Flags().StringVar(&suppressionsCreateEmail, "email", "", "Recipient email to suppress (required)")
	_ = suppressionsCreateCmd.MarkFlagRequired("email")

	suppressionsDeleteCmd.Flags().StringVar(&suppressionsDeleteEmail, "email", "", "Recipient email to delete from suppressions / reactivate (required)")
	_ = suppressionsDeleteCmd.MarkFlagRequired("email")
}

var suppressionsCmd = &cobra.Command{
	Use:   "suppressions",
	Short: "Manage suppressed recipients per message stream",
}

var suppressionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List suppressed recipients for a message stream",
	Long: `List the suppression dump for a message stream — recipients Postmark will
not send to (hard bounces, spam complaints, or manual suppressions).

Examples:
  pmx suppressions list
  pmx suppressions list --stream broadcast
  pmx suppressions list --jq '#.{email:EmailAddress,reason:SuppressionReason}'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		data, err := serverGet(c, "message-streams/"+suppressionsStream+"/suppressions/dump", nil)
		if err != nil {
			return err
		}
		return printData("suppressions.list", extractList(data, "Suppressions"))
	},
}

var suppressionsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Suppress a recipient on a message stream",
	Long: `Add a recipient to a message stream's suppression list so Postmark stops
sending to them on that stream.

Examples:
  pmx suppressions create --email customer@example.com --yes
  pmx suppressions create --email customer@example.com --stream broadcast --dry-run`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if handled, err := gateWrite("suppress " + suppressionsCreateEmail + " on stream " + suppressionsStream); handled {
			return err
		}
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		body := map[string]any{
			"Suppressions": []map[string]any{
				{"EmailAddress": suppressionsCreateEmail},
			},
		}
		data, err := c.Post(ctx(), client.ScopeServer, "message-streams/"+suppressionsStream+"/suppressions", body)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}

var suppressionsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a suppression (reactivate a recipient) on a message stream",
	Long: `Remove a recipient from a message stream's suppression list, reactivating
delivery to them on that stream.

Examples:
  pmx suppressions delete --email customer@example.com --yes
  pmx suppressions delete --email customer@example.com --stream broadcast --dry-run`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if handled, err := gateWrite("delete suppression " + suppressionsDeleteEmail + " (reactivate)"); handled {
			return err
		}
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		body := map[string]any{
			"Suppressions": []map[string]any{
				{"EmailAddress": suppressionsDeleteEmail},
			},
		}
		data, err := c.Post(ctx(), client.ScopeServer, "message-streams/"+suppressionsStream+"/suppressions/delete", body)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}
