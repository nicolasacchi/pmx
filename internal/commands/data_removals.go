package commands

import (
	"net/url"

	"github.com/nicolasacchi/pmx/internal/client"
	"github.com/spf13/cobra"
)

// data_removals.go — GDPR data removals (account scope). Postmark's
// /data-removals endpoints let you request the deletion of all data tied to a
// recipient and poll the resulting request's status.

var dataRemovalsFile string

func init() {
	rootCmd.AddCommand(dataRemovalsCmd)
	dataRemovalsCmd.AddCommand(dataRemovalsCreateCmd, dataRemovalsStatusCmd)

	dataRemovalsCreateCmd.Flags().StringVar(&dataRemovalsFile, "file", "", "JSON body file with RequestedBy/RequestedByEmail/RecipientEmailAddress/NotifyWhenCompleted (\"-\" reads stdin)")
	_ = dataRemovalsCreateCmd.MarkFlagRequired("file")
}

var dataRemovalsCmd = &cobra.Command{
	Use:   "data-removals",
	Short: "Request and track GDPR data removals (account scope)",
}

var dataRemovalsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Request a GDPR data removal for a recipient",
	Long: `Submit a GDPR data removal request (account scope). The --file body is a
JSON document with RequestedBy, RequestedByEmail, RecipientEmailAddress, and
NotifyWhenCompleted. Postmark returns a request ID you can poll with
'pmx data-removals status <id>'.

Examples:
  pmx data-removals create --file removal.json --yes
  echo '{"RequestedBy":"ops@example.com","RequestedByEmail":"ops@example.com","RecipientEmailAddress":"customer@example.com","NotifyWhenCompleted":true}' | pmx data-removals create --file - --yes
  pmx data-removals create --file removal.json --dry-run`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONFile(dataRemovalsFile)
		if err != nil {
			return err
		}
		if handled, err := gateWrite("request data removal from " + dataRemovalsFile); handled {
			return err
		}
		data, err := c.Post(ctx(), client.ScopeAccount, "data-removals", body)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}

var dataRemovalsStatusCmd = &cobra.Command{
	Use:   "status <id>",
	Short: "Check the status of a data removal request",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		data, err := accountGet(c, "data-removals/"+url.PathEscape(args[0]), nil)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}
