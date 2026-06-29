package commands

import (
	"net/url"

	"github.com/nicolasacchi/pmx/internal/client"
	"github.com/spf13/cobra"
)

var (
	templatesCreateFile   string
	templatesEditFile     string
	templatesValidateFile string
)

func init() {
	rootCmd.AddCommand(templatesCmd)
	templatesCmd.AddCommand(
		templatesListCmd,
		templatesGetCmd,
		templatesCreateCmd,
		templatesEditCmd,
		templatesDeleteCmd,
		templatesValidateCmd,
	)

	templatesCreateCmd.Flags().StringVar(&templatesCreateFile, "file", "", "Path to JSON template body (\"-\" reads stdin) (required)")
	_ = templatesCreateCmd.MarkFlagRequired("file")

	templatesEditCmd.Flags().StringVar(&templatesEditFile, "file", "", "Path to JSON template body (\"-\" reads stdin) (required)")
	_ = templatesEditCmd.MarkFlagRequired("file")

	templatesValidateCmd.Flags().StringVar(&templatesValidateFile, "file", "", "Path to JSON template body (\"-\" reads stdin) (required)")
	_ = templatesValidateCmd.MarkFlagRequired("file")
}

var templatesCmd = &cobra.Command{
	Use:   "templates",
	Short: "Manage email templates and layouts",
}

var templatesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List templates and layouts on the server",
	Long: `List all templates (and layouts) on the server.

Examples:
  pmx templates list
  pmx templates list --limit 100
  pmx templates list --jq '#.{id:TemplateId,alias:Alias,name:Name}'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		data, err := serverGet(c, "templates", pageParams())
		if err != nil {
			return err
		}
		return printData("templates.list", extractList(data, "Templates"))
	},
}

var templatesGetCmd = &cobra.Command{
	Use:   "get <idOrAlias>",
	Short: "Get one template by id or alias",
	Long: `Get the full template (subject, HTML/text bodies, associated layout).

Examples:
  pmx templates get 1234567
  pmx templates get welcome-email`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		data, err := serverGet(c, "templates/"+url.PathEscape(args[0]), nil)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}

var templatesCreateCmd = &cobra.Command{
	Use:   "create --file <body.json>",
	Short: "Create a template from a JSON body file",
	Long: `Create a template (or layout) from a JSON request body.

The body is sent verbatim to Postmark; see the Postmark "Create template" API
for the field shape (Name, Alias, Subject, HtmlBody, TextBody, TemplateType, ...).

Examples:
  pmx templates create --file welcome.json
  cat welcome.json | pmx templates create --file - --yes`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONFile(templatesCreateFile)
		if err != nil {
			return err
		}
		if handled, err := gateWrite("create template from " + templatesCreateFile); handled {
			return err
		}
		data, err := c.Post(ctx(), client.ScopeServer, "templates", body)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}

var templatesEditCmd = &cobra.Command{
	Use:   "edit <id> --file <body.json>",
	Short: "Update a template from a JSON body file",
	Long: `Update an existing template (or layout) by id or alias from a JSON body.

Examples:
  pmx templates edit 1234567 --file welcome.json
  pmx templates edit welcome-email --file welcome.json --yes`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONFile(templatesEditFile)
		if err != nil {
			return err
		}
		if handled, err := gateWrite("edit template " + args[0]); handled {
			return err
		}
		data, err := c.Put(ctx(), client.ScopeServer, "templates/"+url.PathEscape(args[0]), body)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}

var templatesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a template by id or alias",
	Long: `Delete a template (or layout). Destructive — requires --yes (or --confirm).

Examples:
  pmx templates delete 1234567 --yes
  pmx templates delete welcome-email --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		if handled, err := gateWrite("delete template " + args[0]); handled {
			return err
		}
		data, err := c.Delete(ctx(), client.ScopeServer, "templates/"+url.PathEscape(args[0]))
		if err != nil {
			return err
		}
		return printData("", data)
	},
}

var templatesValidateCmd = &cobra.Command{
	Use:   "validate --file <body.json>",
	Short: "Validate/preview a template body (no changes made)",
	Long: `Validate a template body and render a preview against test data. This is a
read/preview call — it does NOT create or modify anything, so no confirmation
is required.

The body shape matches the Postmark "Validate template" API (Subject, HtmlBody,
TextBody, TestRenderModel, ...).

Examples:
  pmx templates validate --file draft.json
  cat draft.json | pmx templates validate --file -`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONFile(templatesValidateFile)
		if err != nil {
			return err
		}
		data, err := c.Post(ctx(), client.ScopeServer, "templates/validate", body)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}
