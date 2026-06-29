package commands

import (
	"github.com/nicolasacchi/pmx/internal/client"
	"github.com/spf13/cobra"
)

var serverFile string

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.AddCommand(serverGetCmd, serverEditCmd)

	serverEditCmd.Flags().StringVar(&serverFile, "file", "", "Path to a JSON body with the settings to update (\"-\" reads stdin) (required)")
	_ = serverEditCmd.MarkFlagRequired("file")
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Get or edit the current server's settings",
}

var serverGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get the current server's settings",
	Long: `Get the settings for the current server (the one your server token belongs to):
name, color, inbound/outbound stream config, open/link tracking, bounce hooks.

Examples:
  pmx server get
  pmx server get --jq 'TrackOpens'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		data, err := serverGet(c, "server", nil)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}

var serverEditCmd = &cobra.Command{
	Use:   "edit --file <body.json>",
	Short: "Edit the current server's settings",
	Long: `Update the current server's settings from a JSON body (PATCH /server).
Only the fields present in the body are changed.

Examples:
  pmx server edit --file settings.json --yes
  echo '{"TrackOpens":true}' | pmx server edit --file - --yes
  pmx server edit --file settings.json --dry-run`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if handled, err := gateWrite("edit current server settings"); handled {
			return err
		}
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONFile(serverFile)
		if err != nil {
			return err
		}
		data, err := c.Patch(ctx(), client.ScopeServer, "server", body)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}
