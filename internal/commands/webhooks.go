package commands

import (
	"fmt"
	"net/url"

	"github.com/nicolasacchi/pmx/internal/client"
	"github.com/spf13/cobra"
)

var webhooksStream string

func init() {
	rootCmd.AddCommand(webhooksCmd)
	webhooksCmd.AddCommand(
		webhooksListCmd,
		webhooksGetCmd,
		webhooksCreateCmd,
		webhooksEditCmd,
		webhooksDeleteCmd,
	)

	webhooksListCmd.Flags().StringVar(&webhooksStream, "stream", "", "Filter by message stream id (e.g. outbound, inbound)")

	webhooksCreateCmd.Flags().StringVar(&webhooksFile, "file", "", "JSON request body file ('-' reads stdin)")
	_ = webhooksCreateCmd.MarkFlagRequired("file")

	webhooksEditCmd.Flags().StringVar(&webhooksFile, "file", "", "JSON request body file ('-' reads stdin)")
	_ = webhooksEditCmd.MarkFlagRequired("file")
}

var webhooksFile string

var webhooksCmd = &cobra.Command{
	Use:   "webhooks",
	Short: "Manage webhook subscriptions",
}

var webhooksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List webhook subscriptions",
	Long: `List the server's webhook subscriptions, optionally scoped to one message stream.

Examples:
  pmx webhooks list
  pmx webhooks list --stream outbound
  pmx webhooks list --jq '#.{id:ID,url:Url}'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		params := url.Values{}
		setIfNonEmpty(params, "MessageStream", webhooksStream)

		data, err := serverGet(c, "webhooks", params)
		if err != nil {
			return err
		}
		return printData("webhooks.list", extractList(data, "Webhooks"))
	},
}

var webhooksGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get one webhook subscription",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		data, err := serverGet(c, "webhooks/"+url.PathEscape(args[0]), nil)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}

var webhooksCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a webhook subscription",
	Long: `Create a webhook subscription from a JSON body file.

The body mirrors Postmark's create-webhook payload (Url, MessageStream,
HttpAuth, HttpHeaders, Triggers).

Examples:
  pmx webhooks create --file webhook.json --yes
  pmx webhooks create --file - --dry-run < webhook.json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONFile(webhooksFile)
		if err != nil {
			return err
		}
		if handled, err := gateWrite("create webhook"); handled {
			return err
		}
		data, err := c.Post(ctx(), client.ScopeServer, "webhooks", body)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}

var webhooksEditCmd = &cobra.Command{
	Use:   "edit <id>",
	Short: "Edit a webhook subscription",
	Long: `Edit a webhook subscription from a JSON body file.

Examples:
  pmx webhooks edit 123456 --file webhook.json --yes
  pmx webhooks edit 123456 --file - --dry-run < webhook.json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONFile(webhooksFile)
		if err != nil {
			return err
		}
		if handled, err := gateWrite("edit webhook " + args[0]); handled {
			return err
		}
		data, err := c.Put(ctx(), client.ScopeServer, "webhooks/"+url.PathEscape(args[0]), body)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}

var webhooksDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a webhook subscription",
	Long: `Delete a webhook subscription by id.

Examples:
  pmx webhooks delete 123456 --yes
  pmx webhooks delete 123456 --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		if handled, err := gateWrite(fmt.Sprintf("delete webhook %s", args[0])); handled {
			return err
		}
		data, err := c.Delete(ctx(), client.ScopeServer, "webhooks/"+url.PathEscape(args[0]))
		if err != nil {
			return err
		}
		return printData("", data)
	},
}
