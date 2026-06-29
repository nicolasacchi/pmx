package commands

import (
	"net/url"

	"github.com/nicolasacchi/pmx/internal/client"
	"github.com/spf13/cobra"
)

// streams.go — message streams (server scope). A Postmark "server" holds one or
// more message streams, each typed transactional, broadcast, or inbound; this
// group lists/inspects them and edits/archives the editable ones.

var (
	streamsCreateFile string
	streamsEditFile   string
)

func init() {
	rootCmd.AddCommand(streamsCmd)
	streamsCmd.AddCommand(streamsListCmd, streamsGetCmd, streamsCreateCmd, streamsEditCmd, streamsArchiveCmd)

	cf := streamsCreateCmd.Flags()
	cf.StringVar(&streamsCreateFile, "file", "", "JSON request body file ('-' for stdin) (required)")
	_ = streamsCreateCmd.MarkFlagRequired("file")

	ef := streamsEditCmd.Flags()
	ef.StringVar(&streamsEditFile, "file", "", "JSON request body file ('-' for stdin) (required)")
	_ = streamsEditCmd.MarkFlagRequired("file")
}

var streamsCmd = &cobra.Command{
	Use:   "streams",
	Short: "Manage message streams (transactional vs broadcast)",
}

var streamsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the message streams on this server",
	Long: `List every message stream on the active server, with its type
(Transactional, Broadcasts, or Inbound) and subscription state.

Examples:
  pmx streams list
  pmx streams list --jq '#.{id:ID,type:MessageStreamType,name:Name}'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		data, err := serverGet(c, "message-streams", nil)
		if err != nil {
			return err
		}
		return printData("streams.list", extractList(data, "MessageStreams"))
	},
}

var streamsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get full details for one message stream",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		data, err := serverGet(c, "message-streams/"+url.PathEscape(args[0]), nil)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}

var streamsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new message stream from a JSON body",
	Long: `Create a message stream. The --file body must include at least ID,
Name, and MessageStreamType (Transactional or Broadcasts).

Examples:
  pmx streams create --file new-stream.json --yes
  echo '{"ID":"announcements","Name":"Announcements","MessageStreamType":"Broadcasts"}' | pmx streams create --file - --yes`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if handled, err := gateWrite("create message stream"); handled {
			return err
		}
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONFile(streamsCreateFile)
		if err != nil {
			return err
		}
		data, err := c.Post(ctx(), client.ScopeServer, "message-streams", body)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}

var streamsEditCmd = &cobra.Command{
	Use:   "edit <id>",
	Short: "Edit an existing message stream from a JSON body",
	Long: `Update the editable fields (Name, Description) of a message stream.

Examples:
  pmx streams edit broadcast --file edit.json --yes
  echo '{"Name":"Renamed"}' | pmx streams edit broadcast --file - --yes`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if handled, err := gateWrite("edit message stream " + args[0]); handled {
			return err
		}
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONFile(streamsEditFile)
		if err != nil {
			return err
		}
		data, err := c.Patch(ctx(), client.ScopeServer, "message-streams/"+url.PathEscape(args[0]), body)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}

var streamsArchiveCmd = &cobra.Command{
	Use:   "archive <id>",
	Short: "Archive a message stream (begins the deletion grace period)",
	Long: `Archive a message stream. Postmark schedules it for deletion after a
grace period; until then it can be unarchived from the Postmark UI.

Examples:
  pmx streams archive old-broadcasts --yes
  pmx streams archive old-broadcasts --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if handled, err := gateWrite("archive message stream " + args[0]); handled {
			return err
		}
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		data, err := c.Post(ctx(), client.ScopeServer, "message-streams/"+url.PathEscape(args[0])+"/archive", nil)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}
