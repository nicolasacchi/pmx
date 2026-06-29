package commands

import (
	"net/url"

	"github.com/nicolasacchi/pmx/internal/client"
	"github.com/spf13/cobra"
)

// signatures.go — sender signatures (ACCOUNT scope). Sender signatures back the
// From: addresses Postmark is allowed to send on behalf of, and carry the
// DKIM/SPF/return-path verification state. All endpoints live under /senders and
// require the X-Postmark-Account-Token (account-wide admin) scope.

var (
	signaturesFile string
)

func init() {
	rootCmd.AddCommand(signaturesCmd)
	signaturesCmd.AddCommand(
		signaturesListCmd,
		signaturesGetCmd,
		signaturesCreateCmd,
		signaturesEditCmd,
		signaturesDeleteCmd,
		signaturesResendCmd,
		signaturesRequestDKIMCmd,
	)

	signaturesCreateCmd.Flags().StringVar(&signaturesFile, "file", "", "JSON body file ('-' for stdin)")
	signaturesEditCmd.Flags().StringVar(&signaturesFile, "file", "", "JSON body file ('-' for stdin)")
}

var signaturesCmd = &cobra.Command{
	Use:   "signatures",
	Short: "Manage sender signatures (account scope)",
}

var signaturesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List sender signatures and their DKIM/SPF status",
	Long: `List all sender signatures on the account, including their confirmation and
DKIM/SPF/return-path verification state.

Examples:
  pmx signatures list
  pmx signatures list --jq '#.{id:ID,email:EmailAddress,confirmed:Confirmed}'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		data, err := accountGet(c, "senders", pageParams())
		if err != nil {
			return err
		}
		return printData("signatures.list", extractList(data, "SenderSignatures"))
	},
}

var signaturesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get full details for one sender signature",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		data, err := accountGet(c, "senders/"+url.PathEscape(args[0]), nil)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}

var signaturesCreateCmd = &cobra.Command{
	Use:   "create --file <body.json>",
	Short: "Create a new sender signature",
	Long: `Create a sender signature from a JSON body file ('-' reads stdin). Postmark
sends a confirmation email to the signature's address.

Example:
  echo '{"FromEmail":"ops@example.com","Name":"Ops"}' | pmx signatures create --file - --yes`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if handled, err := gateWrite("create sender signature"); handled {
			return err
		}
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONFile(signaturesFile)
		if err != nil {
			return err
		}
		data, err := c.Post(ctx(), client.ScopeAccount, "senders", body)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}

var signaturesEditCmd = &cobra.Command{
	Use:   "edit <id> --file <body.json>",
	Short: "Update an existing sender signature",
	Long: `Update a sender signature from a JSON body file ('-' reads stdin).

Example:
  echo '{"Name":"Operations"}' | pmx signatures edit 12345 --file - --yes`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if handled, err := gateWrite("edit sender signature " + args[0]); handled {
			return err
		}
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONFile(signaturesFile)
		if err != nil {
			return err
		}
		data, err := c.Put(ctx(), client.ScopeAccount, "senders/"+url.PathEscape(args[0]), body)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}

var signaturesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a sender signature",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if handled, err := gateWrite("delete sender signature " + args[0]); handled {
			return err
		}
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		data, err := c.Delete(ctx(), client.ScopeAccount, "senders/"+url.PathEscape(args[0]))
		if err != nil {
			return err
		}
		return printData("", data)
	},
}

var signaturesResendCmd = &cobra.Command{
	Use:   "resend <id>",
	Short: "Resend the confirmation email for a sender signature",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if handled, err := gateWrite("resend confirmation for signature " + args[0]); handled {
			return err
		}
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		data, err := c.Post(ctx(), client.ScopeAccount, "senders/"+url.PathEscape(args[0])+"/resend", nil)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}

var signaturesRequestDKIMCmd = &cobra.Command{
	Use:   "request-dkim <id>",
	Short: "Request a new DKIM key for a sender signature",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if handled, err := gateWrite("request new DKIM for signature " + args[0]); handled {
			return err
		}
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		data, err := c.Post(ctx(), client.ScopeAccount, "senders/"+url.PathEscape(args[0])+"/requestnewdkim", nil)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}
