package commands

import (
	"github.com/nicolasacchi/pmx/internal/client"
	"github.com/spf13/cobra"
)

var inboundRulesRule string

func init() {
	rootCmd.AddCommand(inboundRulesCmd)
	inboundRulesCmd.AddCommand(inboundRulesListCmd, inboundRulesCreateCmd, inboundRulesDeleteCmd)

	inboundRulesCreateCmd.Flags().StringVar(&inboundRulesRule, "rule", "", "Email address or domain to block (required)")
	_ = inboundRulesCreateCmd.MarkFlagRequired("rule")
}

var inboundRulesCmd = &cobra.Command{
	Use:   "inbound-rules",
	Short: "Manage inbound blocking rules",
}

var inboundRulesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List inbound blocking rules",
	Long: `List the inbound rules that block messages from a given email address or
domain before they reach your inbound stream.

Examples:
  pmx inbound-rules list
  pmx inbound-rules list --jq '#.Rule'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		data, err := serverGet(c, "triggers/inboundrules", pageParams())
		if err != nil {
			return err
		}
		return printData("inbound-rules.list", extractList(data, "InboundRules"))
	},
}

var inboundRulesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an inbound blocking rule",
	Long: `Create an inbound rule that blocks messages from an email address or domain.

Examples:
  pmx inbound-rules create --rule spammer@example.com --yes
  pmx inbound-rules create --rule baddomain.com --dry-run`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		if handled, err := gateWrite("create inbound rule " + inboundRulesRule); handled {
			return err
		}
		data, err := c.Post(ctx(), client.ScopeServer, "triggers/inboundrules", map[string]any{"Rule": inboundRulesRule})
		if err != nil {
			return err
		}
		return printData("", data)
	},
}

var inboundRulesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete an inbound blocking rule",
	Long: `Delete an inbound rule by its numeric id (see 'pmx inbound-rules list').

Examples:
  pmx inbound-rules delete 12345 --yes
  pmx inbound-rules delete 12345 --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		if handled, err := gateWrite("delete inbound rule " + args[0]); handled {
			return err
		}
		data, err := c.Delete(ctx(), client.ScopeServer, "triggers/inboundrules/"+args[0])
		if err != nil {
			return err
		}
		return printData("", data)
	},
}
