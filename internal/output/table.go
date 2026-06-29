package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"golang.org/x/term"
)

type FormatFunc func(any) string

type ColumnDef struct {
	Header string
	Key    string // Postmark JSON keys are PascalCase — match exactly
	Format FormatFunc
}

// commandColumns defines the human (TTY) table for each command. A command with
// no entry here always renders JSON (printTable returns an error → JSON path).
var commandColumns = map[string][]ColumnDef{
	"messages.outbound": {
		{Header: "MESSAGEID", Key: "MessageID"},
		{Header: "TO", Key: "Recipients"},
		{Header: "SUBJECT", Key: "Subject", Format: truncate60},
		{Header: "STATUS", Key: "Status"},
		{Header: "RECEIVED", Key: "ReceivedAt"},
	},
	"messages.inbound": {
		{Header: "MESSAGEID", Key: "MessageID"},
		{Header: "FROM", Key: "From"},
		{Header: "SUBJECT", Key: "Subject", Format: truncate60},
		{Header: "STATUS", Key: "Status"},
		{Header: "RECEIVED", Key: "ReceivedAt"},
	},
	"bounces.list": {
		{Header: "ID", Key: "ID"},
		{Header: "EMAIL", Key: "Email"},
		{Header: "TYPE", Key: "Type"},
		{Header: "INACTIVE", Key: "Inactive"},
		{Header: "BOUNCEDAT", Key: "BouncedAt"},
	},
	"suppressions.list": {
		{Header: "EMAIL", Key: "EmailAddress"},
		{Header: "REASON", Key: "SuppressionReason"},
		{Header: "ORIGIN", Key: "Origin"},
		{Header: "CREATED", Key: "CreatedAt"},
	},
	"stats.list": {
		{Header: "DATE", Key: "Date"},
		{Header: "SENT", Key: "Sent"},
	},
	"templates.list": {
		{Header: "ID", Key: "TemplateId"},
		{Header: "ALIAS", Key: "Alias"},
		{Header: "NAME", Key: "Name"},
		{Header: "ACTIVE", Key: "Active"},
		{Header: "TYPE", Key: "TemplateType"},
	},
	"streams.list": {
		{Header: "ID", Key: "ID"},
		{Header: "NAME", Key: "Name"},
		{Header: "TYPE", Key: "MessageStreamType"},
	},
	"webhooks.list": {
		{Header: "ID", Key: "ID"},
		{Header: "URL", Key: "Url", Format: truncate60},
		{Header: "STREAM", Key: "MessageStream"},
	},
	"inbound-rules.list": {
		{Header: "ID", Key: "ID"},
		{Header: "RULE", Key: "Rule"},
	},
	"servers.list": {
		{Header: "ID", Key: "ID"},
		{Header: "NAME", Key: "Name"},
		{Header: "COLOR", Key: "Color"},
	},
	"signatures.list": {
		{Header: "ID", Key: "ID"},
		{Header: "EMAIL", Key: "EmailAddress"},
		{Header: "CONFIRMED", Key: "Confirmed"},
		{Header: "DKIM", Key: "DKIMVerified"},
	},
	"domains.list": {
		{Header: "ID", Key: "ID"},
		{Header: "NAME", Key: "Name"},
		{Header: "DKIM", Key: "DKIMVerified"},
		{Header: "RETURNPATH", Key: "ReturnPathDomainVerified"},
	},
	"config.list": {
		{Header: "NAME", Key: "name"},
		{Header: "SERVER TOKEN", Key: "server_token"},
		{Header: "ACCOUNT TOKEN", Key: "account_token"},
		{Header: "DEFAULT", Key: "default"},
	},
}

func printTable(command string, data json.RawMessage) error {
	columns, ok := commandColumns[command]
	if !ok {
		return fmt.Errorf("no table definition for %s", command)
	}

	var rows []map[string]any
	if err := json.Unmarshal(data, &rows); err != nil {
		var single map[string]any
		if err2 := json.Unmarshal(data, &single); err2 != nil {
			return fmt.Errorf("cannot render as table")
		}
		rows = []map[string]any{single}
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	if term.IsTerminal(int(os.Stdout.Fd())) {
		t.SetStyle(table.StyleLight)
	} else {
		t.SetStyle(table.StyleDefault)
	}

	header := make(table.Row, len(columns))
	for i, col := range columns {
		header[i] = col.Header
	}
	t.AppendHeader(header)

	for _, row := range rows {
		r := make(table.Row, len(columns))
		for i, col := range columns {
			if col.Format != nil {
				r[i] = col.Format(row[col.Key])
			} else {
				r[i] = formatValue(row[col.Key])
			}
		}
		t.AppendRow(r)
	}

	t.Render()
	return nil
}

func formatValue(v any) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%.2f", val)
	case bool:
		if val {
			return "true"
		}
		return "false"
	case []any:
		parts := make([]string, len(val))
		for i, item := range val {
			parts[i] = fmt.Sprintf("%v", item)
		}
		return strings.Join(parts, ", ")
	default:
		b, _ := json.Marshal(val)
		return string(b)
	}
}

func truncate60(v any) string {
	s := formatValue(v)
	if len(s) > 60 {
		return s[:57] + "..."
	}
	return s
}
