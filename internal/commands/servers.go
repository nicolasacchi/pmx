package commands

import (
	"net/url"

	"github.com/nicolasacchi/pmx/internal/client"
	"github.com/spf13/cobra"
)

// servers is the ACCOUNT-scope plural group: manage every Postmark "server"
// across the account (list/get/create/edit/delete). Reads use accountGet
// (X-Postmark-Account-Token); writes go through c.Post/Put/Delete on
// client.ScopeAccount behind gateWrite.

var (
	serversName string
	serversFile string
)

func init() {
	rootCmd.AddCommand(serversCmd)
	serversCmd.AddCommand(serversListCmd, serversGetCmd, serversCreateCmd, serversEditCmd, serversDeleteCmd)

	serversListCmd.Flags().StringVar(&serversName, "name", "", "Filter by server name")
	serversCreateCmd.Flags().StringVar(&serversFile, "file", "", "Path to a JSON body file ('-' reads stdin)")
	serversEditCmd.Flags().StringVar(&serversFile, "file", "", "Path to a JSON body file ('-' reads stdin)")
}

var serversCmd = &cobra.Command{
	Use:   "servers",
	Short: "Manage servers across the account (account scope)",
}

var serversListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all servers on the account",
	Long: `List every Postmark server on the account (account scope).

Examples:
  pmx servers list
  pmx servers list --name production
  pmx servers list --jq '#.{id:ID,name:Name,color:Color}'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		params := pageParams()
		setIfNonEmpty(params, "name", serversName)

		data, err := accountGet(c, "servers", params)
		if err != nil {
			return err
		}
		return printData("servers.list", extractList(data, "Servers"))
	},
}

var serversGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get full details for one server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		data, err := accountGet(c, "servers/"+url.PathEscape(args[0]), nil)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}

var serversCreateCmd = &cobra.Command{
	Use:   "create --file <body.json>",
	Short: "Create a new server (write)",
	Long: `Create a new Postmark server from a JSON body file (account scope).

The body mirrors the Postmark API (at minimum a "Name"); '-' reads stdin.

Examples:
  pmx servers create --file new-server.json --yes
  echo '{"Name":"staging","Color":"blue"}' | pmx servers create --file - --yes`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if handled, err := gateWrite("create server"); handled {
			return err
		}
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONFile(serversFile)
		if err != nil {
			return err
		}
		data, err := c.Post(ctx(), client.ScopeAccount, "servers", body)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}

var serversEditCmd = &cobra.Command{
	Use:   "edit <id> --file <body.json>",
	Short: "Edit an existing server (write)",
	Long: `Update a Postmark server from a JSON body file (account scope).

Only the fields present in the body are changed; '-' reads stdin.

Examples:
  pmx servers edit 12345 --file changes.json --yes
  echo '{"Color":"red"}' | pmx servers edit 12345 --file - --yes`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		if handled, err := gateWrite("edit server " + id); handled {
			return err
		}
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONFile(serversFile)
		if err != nil {
			return err
		}
		data, err := c.Put(ctx(), client.ScopeAccount, "servers/"+url.PathEscape(id), body)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}

var serversDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a server (write, irreversible)",
	Long: `Delete a Postmark server (account scope). This is irreversible.

Example:
  pmx servers delete 12345 --yes`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		if handled, err := gateWrite("delete server " + id); handled {
			return err
		}
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		data, err := c.Delete(ctx(), client.ScopeAccount, "servers/"+url.PathEscape(id))
		if err != nil {
			return err
		}
		return printData("", data)
	},
}
