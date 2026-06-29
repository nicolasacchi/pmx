package commands

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/nicolasacchi/pmx/internal/client"
	"github.com/nicolasacchi/pmx/internal/config"
	"github.com/spf13/cobra"
)

var (
	configServerToken  string
	configAccountToken string
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configAddCmd, configRemoveCmd, configListCmd, configUseCmd, configCurrentCmd, configDoctorCmd)

	configAddCmd.Flags().StringVar(&configServerToken, "server-token", "", "Postmark Server token")
	configAddCmd.Flags().StringVar(&configAccountToken, "account-token", "", "Postmark Account token")
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage pmx configuration (~/.config/pmx/config.toml)",
}

var configAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add or update a project (server/account tokens)",
	Long: `Add a named project. Provide at least one token scope.

  pmx config add production --server-token TOKEN --account-token TOKEN
  pmx config add staging    --server-token TOKEN`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if configServerToken == "" && configAccountToken == "" {
			return fmt.Errorf("provide at least one of --server-token / --account-token")
		}
		if err := config.AddProject(args[0], configServerToken, configAccountToken); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Project %q saved\n", args[0])
		return nil
	},
}

var configRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.RemoveProject(args[0]); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Project %q removed\n", args[0])
		return nil
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.ListProjects()
		if err != nil {
			return fmt.Errorf("no config file found — run 'pmx config add' to create one")
		}
		if len(cfg.Projects) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No projects configured")
			return nil
		}
		var rows []map[string]any
		for name, p := range cfg.Projects {
			rows = append(rows, map[string]any{
				"name":          name,
				"server_token":  config.MaskKey(p.ServerToken),
				"account_token": config.MaskKey(p.AccountToken),
				"default":       name == cfg.DefaultProject,
			})
		}
		data, _ := json.Marshal(rows)
		return printData("config.list", data)
	},
}

var configUseCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Set the default project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.SetDefaultProject(args[0]); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Default project set to %q\n", args[0])
		return nil
	},
}

var configCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show the current default project",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.ListProjects()
		if err != nil {
			return fmt.Errorf("no config file found")
		}
		if cfg.DefaultProject == "" {
			fmt.Fprintln(cmd.OutOrStdout(), "No default project set")
			return nil
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Current project: %s\n", cfg.DefaultProject)
		return nil
	},
}

var configDoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Probe each token scope and report what it can reach",
	Long: `Resolve the configured tokens and probe one cheap endpoint per scope:
  server  → GET /server         (the current server's settings)
  account → GET /servers?count=1 (account-wide server list)
so a wrong or missing token fails loud rather than returning empty results.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		report := map[string]any{}

		// Server scope probe.
		if _, err := c.Get(ctx(), client.ScopeServer, "server", nil); err != nil {
			report["server_scope"] = map[string]any{"ok": false, "error": err.Error()}
		} else {
			report["server_scope"] = map[string]any{"ok": true}
		}

		// Account scope probe.
		ap := url.Values{}
		ap.Set("count", "1")
		ap.Set("offset", "0")
		if _, err := c.Get(ctx(), client.ScopeAccount, "servers", ap); err != nil {
			report["account_scope"] = map[string]any{"ok": false, "error": err.Error()}
		} else {
			report["account_scope"] = map[string]any{"ok": true}
		}

		data, _ := json.Marshal(report)
		return printData("", data)
	},
}
