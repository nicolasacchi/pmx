package commands

import (
	"encoding/json"
	"net/url"
	"strconv"
	"sync"

	"github.com/nicolasacchi/pmx/internal/client"
	"github.com/spf13/cobra"
)

// otp.go is a composite convenience command for the headline use case: a
// customer says their passwordless sign-in code never arrived. It answers, in
// one call, "what happened to this person's recent mail?" by fanning out:
//   - recent outbound messages to the recipient (status + timeline)
//   - any bounces for the address (a hard bounce silently suppresses future OTP)
// OTP is sent over SMTP but Postmark tracks it identically, so this works even
// though the OTP mailer may not set an X-PM-Tag.

func init() {
	rootCmd.AddCommand(otpCmd)
}

var otpCmd = &cobra.Command{
	Use:   "otp <email>",
	Short: "Trace recent transactional mail (incl. OTP) for a recipient",
	Long: `Diagnose "I never got my code" in one shot (server scope):

  pmx otp customer@example.com
  pmx otp customer@example.com --from 7d --limit 20

Returns recent outbound messages to the address with their delivery status,
the event timeline of the most recent one, and any bounce/suppression that
would silently drop further mail.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		email := args[0]

		count := limitFlag
		if count <= 0 || count > 50 {
			count = 10
		}

		result := map[string]any{"recipient": email}
		var mu sync.Mutex
		var wg sync.WaitGroup

		// Recent outbound messages to this recipient.
		wg.Add(1)
		go func() {
			defer wg.Done()
			params := url.Values{}
			params.Set("recipient", email)
			params.Set("count", strconv.Itoa(count))
			params.Set("offset", "0")
			if from, err := fromDate(); err == nil && from != "" {
				params.Set("fromdate", from)
			}
			data, err := c.Get(ctx(), client.ScopeServer, "messages/outbound", params)
			if err != nil {
				mu.Lock()
				result["messages_error"] = err.Error()
				mu.Unlock()
				return
			}
			msgs := extractList(data, "Messages")
			mu.Lock()
			var raw any
			_ = json.Unmarshal(msgs, &raw)
			result["messages"] = raw
			result["message_count"] = jsonLen(msgs)
			mu.Unlock()

			// Timeline of the most recent message.
			latestID := firstMessageID(msgs)
			if latestID != "" {
				if det, err := c.Get(ctx(), client.ScopeServer, "messages/outbound/"+url.PathEscape(latestID)+"/details", nil); err == nil {
					var d struct {
						MessageEvents json.RawMessage `json:"MessageEvents"`
					}
					if json.Unmarshal(det, &d) == nil && d.MessageEvents != nil {
						var ev any
						_ = json.Unmarshal(d.MessageEvents, &ev)
						mu.Lock()
						result["latest_message_id"] = latestID
						result["latest_events"] = ev
						mu.Unlock()
					}
				}
			}
		}()

		// Bounces for this address (inactive => suppressing future sends).
		wg.Add(1)
		go func() {
			defer wg.Done()
			params := url.Values{}
			params.Set("emailFilter", email)
			params.Set("count", "50")
			params.Set("offset", "0")
			data, err := c.Get(ctx(), client.ScopeServer, "bounces", params)
			if err != nil {
				mu.Lock()
				result["bounces_error"] = err.Error()
				mu.Unlock()
				return
			}
			b := extractList(data, "Bounces")
			mu.Lock()
			var raw any
			_ = json.Unmarshal(b, &raw)
			result["bounces"] = raw
			result["bounce_count"] = jsonLen(b)
			mu.Unlock()
		}()

		wg.Wait()
		out, _ := json.Marshal(result)
		return printData("", out)
	},
}

// firstMessageID returns the MessageID of the first element of a messages array.
func firstMessageID(arr json.RawMessage) string {
	var msgs []struct {
		MessageID string `json:"MessageID"`
	}
	if json.Unmarshal(arr, &msgs) == nil && len(msgs) > 0 {
		return msgs[0].MessageID
	}
	return ""
}
