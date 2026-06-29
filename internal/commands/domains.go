package commands

import (
	"net/url"

	"github.com/nicolasacchi/pmx/internal/client"
	"github.com/spf13/cobra"
)

// domains.go — account-scope group for sending domains and their DKIM/SPF
// (Return-Path) state. Reads use accountGet; writes go through the account
// scope and the gateWrite confirm gate. Mirrors the messages.go exemplar.

var (
	domainsCreateFile string
	domainsEditFile   string
)

func init() {
	rootCmd.AddCommand(domainsCmd)
	domainsCmd.AddCommand(
		domainsListCmd,
		domainsGetCmd,
		domainsCreateCmd,
		domainsEditCmd,
		domainsDeleteCmd,
		domainsVerifyDkimCmd,
		domainsVerifyReturnPathCmd,
		domainsRotateDkimCmd,
	)

	domainsCreateCmd.Flags().StringVar(&domainsCreateFile, "file", "", "JSON request body file ('-' for stdin)")
	domainsEditCmd.Flags().StringVar(&domainsEditFile, "file", "", "JSON request body file ('-' for stdin)")
}

var domainsCmd = &cobra.Command{
	Use:   "domains",
	Short: "Manage sending domains and DKIM/SPF (account scope)",
}

var domainsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List sending domains with DKIM/Return-Path verification state",
	Long: `List the sending domains on the account.

Examples:
  pmx domains list
  pmx domains list --jq '#.{id:ID,name:Name,dkim:DKIMVerified,spf:ReturnPathDomainVerified}'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		data, err := accountGet(c, "domains", pageParams())
		if err != nil {
			return err
		}
		return printData("domains.list", extractList(data, "Domains"))
	},
}

var domainsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get full details for one sending domain (DKIM/SPF records)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		data, err := accountGet(c, "domains/"+url.PathEscape(args[0]), nil)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}

var domainsCreateCmd = &cobra.Command{
	Use:   "create --file <body.json>",
	Short: "Create a sending domain (WRITE)",
	Long: `Create a sending domain from a JSON body file.

Examples:
  pmx domains create --file domain.json --yes
  echo '{"Name":"example.com"}' | pmx domains create --file - --yes`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		if handled, err := gateWrite("create domain"); handled {
			return err
		}
		body, err := readJSONFile(domainsCreateFile)
		if err != nil {
			return err
		}
		data, err := c.Post(ctx(), client.ScopeAccount, "domains", body)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}

var domainsEditCmd = &cobra.Command{
	Use:   "edit <id> --file <body.json>",
	Short: "Edit a sending domain (WRITE)",
	Long: `Update a sending domain's settings from a JSON body file.

Examples:
  pmx domains edit 123456 --file domain.json --yes`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		if handled, err := gateWrite("edit domain " + args[0]); handled {
			return err
		}
		body, err := readJSONFile(domainsEditFile)
		if err != nil {
			return err
		}
		data, err := c.Put(ctx(), client.ScopeAccount, "domains/"+url.PathEscape(args[0]), body)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}

var domainsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a sending domain (WRITE)",
	Long: `Delete a sending domain.

Examples:
  pmx domains delete 123456 --yes
  pmx domains delete 123456 --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		if handled, err := gateWrite("delete domain " + args[0]); handled {
			return err
		}
		data, err := c.Delete(ctx(), client.ScopeAccount, "domains/"+url.PathEscape(args[0]))
		if err != nil {
			return err
		}
		return printData("", data)
	},
}

var domainsVerifyDkimCmd = &cobra.Command{
	Use:   "verify-dkim <id>",
	Short: "Trigger DKIM verification for a sending domain (WRITE)",
	Long: `Ask Postmark to re-check the domain's DKIM DNS record and update its state.

Examples:
  pmx domains verify-dkim 123456 --yes`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		if handled, err := gateWrite("verify DKIM for domain " + args[0]); handled {
			return err
		}
		data, err := c.Put(ctx(), client.ScopeAccount, "domains/"+url.PathEscape(args[0])+"/verifyDkim", nil)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}

var domainsVerifyReturnPathCmd = &cobra.Command{
	Use:   "verify-returnpath <id>",
	Short: "Trigger Return-Path (SPF) verification for a sending domain (WRITE)",
	Long: `Ask Postmark to re-check the domain's Return-Path CNAME and update its state.

Examples:
  pmx domains verify-returnpath 123456 --yes`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		if handled, err := gateWrite("verify Return-Path for domain " + args[0]); handled {
			return err
		}
		data, err := c.Put(ctx(), client.ScopeAccount, "domains/"+url.PathEscape(args[0])+"/verifyReturnPath", nil)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}

var domainsRotateDkimCmd = &cobra.Command{
	Use:   "rotate-dkim <id>",
	Short: "Rotate (regenerate) the DKIM key for a sending domain (WRITE)",
	Long: `Generate a new DKIM key pair for the domain. The new public key must be
published in DNS, then verified with 'pmx domains verify-dkim'.

Examples:
  pmx domains rotate-dkim 123456 --yes`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		if handled, err := gateWrite("rotate DKIM for domain " + args[0]); handled {
			return err
		}
		data, err := c.Post(ctx(), client.ScopeAccount, "domains/"+url.PathEscape(args[0])+"/rotatedkim", nil)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}
