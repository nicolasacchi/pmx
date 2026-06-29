package commands

import (
	"encoding/json"
	"net/url"

	"github.com/spf13/cobra"
)

// messages.go is the EXEMPLAR command group — every other server-scope group
// (bounces, stats, suppressions, templates, streams, server, webhooks,
// inbound-rules) follows this shape:
//   - one file per group, package `commands`
//   - all package-level identifiers prefixed with the group name (no collisions)
//   - register in init() via rootCmd.AddCommand (no edits to root.go)
//   - read via serverGet/accountGet, write via c.Post/Put/Patch/Delete + gateWrite
//   - list commands unwrap the Postmark envelope with extractList(data, "<Key>")
//     and render through printData("group.sub", data)

var (
	messagesRecipient string
	messagesTag       string
	messagesStatus    string
	messagesStream    string
	messagesFromEmail string
	messagesSubject   string
)

func init() {
	rootCmd.AddCommand(messagesCmd)
	messagesCmd.AddCommand(messagesOutboundCmd, messagesInboundCmd)
	messagesOutboundCmd.AddCommand(messagesOutboundGetCmd, messagesOutboundDumpCmd, messagesOutboundEventsCmd)
	messagesInboundCmd.AddCommand(messagesInboundGetCmd, messagesInboundDumpCmd)

	of := messagesOutboundCmd.Flags()
	of.StringVar(&messagesRecipient, "recipient", "", "Filter by recipient email")
	of.StringVar(&messagesFromEmail, "from-email", "", "Filter by sender email (Postmark 'fromemail')")
	of.StringVar(&messagesTag, "tag", "", "Filter by message tag (set via X-PM-Tag; OTP mail may not carry one)")
	of.StringVar(&messagesStatus, "status", "", "Filter by status: queued, sent, processed, delivered, bounced")
	of.StringVar(&messagesStream, "stream", "", "Filter by message stream id (e.g. outbound)")
	of.StringVar(&messagesSubject, "subject", "", "Filter by subject")

	inf := messagesInboundCmd.Flags()
	inf.StringVar(&messagesRecipient, "recipient", "", "Filter by recipient email")
	inf.StringVar(&messagesFromEmail, "from-email", "", "Filter by sender email")
	inf.StringVar(&messagesSubject, "subject", "", "Filter by subject")
}

var messagesCmd = &cobra.Command{
	Use:   "messages",
	Short: "Search outbound/inbound messages and trace delivery (server scope)",
}

// --- outbound -------------------------------------------------------------

var messagesOutboundCmd = &cobra.Command{
	Use:   "outbound",
	Short: "List outbound messages (the OTP delivery lens)",
	Long: `Search outbound messages. Postmark tracks SMTP-sent mail (like the
passwordless OTP) the same as API-sent mail, so this returns OTP traffic too.

Examples:
  pmx messages outbound --recipient customer@example.com --from 2d
  pmx messages outbound --status bounced --from 24h
  pmx messages outbound --tag otp --jq '#.{id:MessageID,status:Status,at:ReceivedAt}'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		params := pageParams()
		if err := addDateParams(params); err != nil {
			return err
		}
		setIfNonEmpty(params, "recipient", messagesRecipient)
		setIfNonEmpty(params, "fromemail", messagesFromEmail)
		setIfNonEmpty(params, "tag", messagesTag)
		setIfNonEmpty(params, "status", messagesStatus)
		setIfNonEmpty(params, "messagestream", messagesStream)
		setIfNonEmpty(params, "subject", messagesSubject)

		data, err := serverGet(c, "messages/outbound", params)
		if err != nil {
			return err
		}
		return printData("messages.outbound", extractList(data, "Messages"))
	},
}

var messagesOutboundGetCmd = &cobra.Command{
	Use:   "get <messageid>",
	Short: "Get full details for one outbound message",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		data, err := serverGet(c, "messages/outbound/"+url.PathEscape(args[0])+"/details", nil)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}

var messagesOutboundDumpCmd = &cobra.Command{
	Use:   "dump <messageid>",
	Short: "Get the raw rendered body of an outbound message",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		data, err := serverGet(c, "messages/outbound/"+url.PathEscape(args[0])+"/dump", nil)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}

var messagesOutboundEventsCmd = &cobra.Command{
	Use:   "events <messageid>",
	Short: "Show the delivery/open/click/bounce event timeline for a message",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		data, err := serverGet(c, "messages/outbound/"+url.PathEscape(args[0])+"/details", nil)
		if err != nil {
			return err
		}
		// MessageEvents is the timeline embedded in the details payload.
		var details struct {
			MessageEvents json.RawMessage `json:"MessageEvents"`
		}
		if json.Unmarshal(data, &details) == nil && details.MessageEvents != nil {
			data = details.MessageEvents
		}
		return printData("", data)
	},
}

// --- inbound --------------------------------------------------------------

var messagesInboundCmd = &cobra.Command{
	Use:   "inbound",
	Short: "List inbound messages",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		params := pageParams()
		if err := addDateParams(params); err != nil {
			return err
		}
		setIfNonEmpty(params, "recipient", messagesRecipient)
		setIfNonEmpty(params, "fromemail", messagesFromEmail)
		setIfNonEmpty(params, "subject", messagesSubject)

		data, err := serverGet(c, "messages/inbound", params)
		if err != nil {
			return err
		}
		return printData("messages.inbound", extractList(data, "InboundMessages"))
	},
}

var messagesInboundGetCmd = &cobra.Command{
	Use:   "get <messageid>",
	Short: "Get full details for one inbound message",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		data, err := serverGet(c, "messages/inbound/"+url.PathEscape(args[0])+"/details", nil)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}

var messagesInboundDumpCmd = &cobra.Command{
	Use:   "dump <messageid>",
	Short: "Get the raw body of an inbound message",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		data, err := serverGet(c, "messages/inbound/"+url.PathEscape(args[0])+"/dump", nil)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}

// setIfNonEmpty sets a query param only when the value is non-empty.
func setIfNonEmpty(p url.Values, key, value string) {
	if value != "" {
		p.Set(key, value)
	}
}
