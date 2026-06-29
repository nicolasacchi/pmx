package commands

import (
	"fmt"

	"github.com/nicolasacchi/pmx/internal/client"
	"github.com/spf13/cobra"
)

// email is a server-scope group for sending transactional mail. It exists for
// testing / smoke checks (e.g. verifying the OTP path delivers) rather than
// production sending. Every leaf is a WRITE and runs through gateWrite.

var (
	emailFrom       string
	emailTo         string
	emailCc         string
	emailBcc        string
	emailSubject    string
	emailText       string
	emailHTML       string
	emailTag        string
	emailStream     string
	emailReplyTo    string
	emailTemplateID int
	emailAlias      string
	emailModelFile  string
	emailFile       string
)

func init() {
	rootCmd.AddCommand(emailCmd)
	emailCmd.AddCommand(emailSendCmd, emailSendTemplateCmd, emailSendBatchCmd)

	sf := emailSendCmd.Flags()
	sf.StringVar(&emailFrom, "from", "", "Sender address (must be a verified signature/domain) [required]")
	sf.StringVar(&emailTo, "to", "", "Recipient address(es), comma-separated [required]")
	sf.StringVar(&emailCc, "cc", "", "Cc address(es), comma-separated")
	sf.StringVar(&emailBcc, "bcc", "", "Bcc address(es), comma-separated")
	sf.StringVar(&emailSubject, "subject", "", "Email subject [required]")
	sf.StringVar(&emailText, "text", "", "Plain-text body (provide --text and/or --html)")
	sf.StringVar(&emailHTML, "html", "", "HTML body (provide --text and/or --html)")
	sf.StringVar(&emailTag, "tag", "", "Message tag for filtering/stats")
	sf.StringVar(&emailStream, "stream", "", "Message stream id (default: outbound)")
	sf.StringVar(&emailReplyTo, "reply-to", "", "Reply-To address")

	tf := emailSendTemplateCmd.Flags()
	tf.StringVar(&emailFrom, "from", "", "Sender address (verified signature/domain) [required]")
	tf.StringVar(&emailTo, "to", "", "Recipient address(es), comma-separated [required]")
	tf.IntVar(&emailTemplateID, "template-id", 0, "Template id (use this OR --alias, not both)")
	tf.StringVar(&emailAlias, "alias", "", "Template alias (use this OR --template-id, not both)")
	tf.StringVar(&emailModelFile, "model-file", "", "JSON file with the template model ('-' = stdin)")
	tf.StringVar(&emailStream, "stream", "", "Message stream id (default: outbound)")
	tf.StringVar(&emailTag, "tag", "", "Message tag for filtering/stats")

	bf := emailSendBatchCmd.Flags()
	bf.StringVar(&emailFile, "file", "", "JSON file containing an array of message objects ('-' = stdin) [required]")
}

var emailCmd = &cobra.Command{
	Use:   "email",
	Short: "Send transactional email (testing / smoke checks)",
}

var emailSendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send a single transactional email",
	Long: `Send one email through Postmark's /email endpoint. Requires --from, --to,
--subject and at least one of --text/--html. This is a WRITE — re-run with --yes
to send, or --dry-run to preview.

Examples:
  pmx email send --from no-reply@example.com --to me@example.com \
    --subject "smoke test" --text "hello from pmx" --yes
  pmx email send --from no-reply@example.com --to me@example.com \
    --subject "smoke test" --html "<p>hi</p>" --tag smoke --dry-run`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if emailFrom == "" || emailTo == "" || emailSubject == "" {
			return fmt.Errorf("--from, --to and --subject are required")
		}
		if emailText == "" && emailHTML == "" {
			return fmt.Errorf("provide at least one of --text or --html")
		}
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		body := map[string]any{}
		emailSetIfNonEmpty(body, "From", emailFrom)
		emailSetIfNonEmpty(body, "To", emailTo)
		emailSetIfNonEmpty(body, "Cc", emailCc)
		emailSetIfNonEmpty(body, "Bcc", emailBcc)
		emailSetIfNonEmpty(body, "Subject", emailSubject)
		emailSetIfNonEmpty(body, "TextBody", emailText)
		emailSetIfNonEmpty(body, "HtmlBody", emailHTML)
		emailSetIfNonEmpty(body, "Tag", emailTag)
		emailSetIfNonEmpty(body, "MessageStream", emailStream)
		emailSetIfNonEmpty(body, "ReplyTo", emailReplyTo)

		if handled, err := gateWrite("send email to " + emailTo); handled {
			return err
		}
		data, err := c.Post(ctx(), client.ScopeServer, "email", body)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}

var emailSendTemplateCmd = &cobra.Command{
	Use:   "send-template",
	Short: "Send a templated email by template id or alias",
	Long: `Send an email rendered from a Postmark template. Requires --from, --to and
exactly one of --template-id / --alias. The template model is read from a JSON
file via --model-file ('-' = stdin). This is a WRITE — re-run with --yes to
send, or --dry-run to preview.

Examples:
  pmx email send-template --from no-reply@example.com --to me@example.com \
    --alias welcome --model-file model.json --yes
  pmx email send-template --from no-reply@example.com --to me@example.com \
    --template-id 12345 --model-file - --dry-run`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if emailFrom == "" || emailTo == "" {
			return fmt.Errorf("--from and --to are required")
		}
		hasID := emailTemplateID > 0
		hasAlias := emailAlias != ""
		if hasID == hasAlias {
			return fmt.Errorf("provide exactly one of --template-id or --alias")
		}
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		body := map[string]any{}
		emailSetIfNonEmpty(body, "From", emailFrom)
		emailSetIfNonEmpty(body, "To", emailTo)
		if hasID {
			body["TemplateId"] = emailTemplateID
		} else {
			body["TemplateAlias"] = emailAlias
		}
		if emailModelFile != "" {
			model, err := readJSONFile(emailModelFile)
			if err != nil {
				return err
			}
			body["TemplateModel"] = model
		}
		emailSetIfNonEmpty(body, "MessageStream", emailStream)
		emailSetIfNonEmpty(body, "Tag", emailTag)

		if handled, err := gateWrite("send template email to " + emailTo); handled {
			return err
		}
		data, err := c.Post(ctx(), client.ScopeServer, "email/withTemplate", body)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}

var emailSendBatchCmd = &cobra.Command{
	Use:   "send-batch",
	Short: "Send a batch of emails from a JSON array file",
	Long: `Send up to 500 messages in one call. --file points at a JSON file containing
an array of message objects (same shape as the /email payload; '-' = stdin).
This is a WRITE — re-run with --yes to send, or --dry-run to preview.

Examples:
  pmx email send-batch --file batch.json --yes
  cat batch.json | pmx email send-batch --file - --dry-run`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if emailFile == "" {
			return fmt.Errorf("--file is required (JSON array of message objects)")
		}
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONFile(emailFile)
		if err != nil {
			return err
		}
		if handled, err := gateWrite("send batch email from " + emailFile); handled {
			return err
		}
		data, err := c.Post(ctx(), client.ScopeServer, "email/batch", body)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}

// emailSetIfNonEmpty adds a string field to the request body only when non-empty
// (group-local to avoid colliding with the query-param setIfNonEmpty helper).
func emailSetIfNonEmpty(body map[string]any, key, value string) {
	if value != "" {
		body[key] = value
	}
}
