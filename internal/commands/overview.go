package commands

import (
	"encoding/json"
	"net/url"
	"sync"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(overviewCmd)
}

var overviewCmd = &cobra.Command{
	Use:   "overview",
	Short: "Parallel server health snapshot: delivery stats, recent bounces, suppressions, streams",
	Long: `Fetch a server's headline health signals in parallel — delivery stats,
recent outbound stats, the count of inactive bounces, and the configured
message streams — and emit them as a single JSON object. Each probe reports
independently: a failing call lands under "<name>_error" instead of aborting
the snapshot.

Examples:
  pmx overview
  pmx overview --jq 'inactive_bounces'
  pmx overview --jq '{bounces:inactive_bounces,streams:streams.#}'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		result := map[string]any{}
		var mu sync.Mutex
		var wg sync.WaitGroup

		// Delivery stats (bounce rate / inactive totals).
		wg.Add(1)
		go func() {
			defer wg.Done()
			data, err := serverGet(c, "deliverystats", nil)
			if err != nil {
				mu.Lock()
				result["delivery_stats_error"] = err.Error()
				mu.Unlock()
				return
			}
			var parsed any
			json.Unmarshal(data, &parsed)
			mu.Lock()
			result["delivery_stats"] = parsed
			mu.Unlock()
		}()

		// Outbound sending stats (sent/opens/bounces over the default window).
		wg.Add(1)
		go func() {
			defer wg.Done()
			data, err := serverGet(c, "stats/outbound", nil)
			if err != nil {
				mu.Lock()
				result["outbound_stats_error"] = err.Error()
				mu.Unlock()
				return
			}
			var parsed any
			json.Unmarshal(data, &parsed)
			mu.Lock()
			result["outbound_stats"] = parsed
			mu.Unlock()
		}()

		// Recent inactive bounces — report just the count for the snapshot.
		wg.Add(1)
		go func() {
			defer wg.Done()
			p := url.Values{}
			p.Set("inactive", "true")
			p.Set("count", "20")
			p.Set("offset", "0")
			data, err := serverGet(c, "bounces", p)
			if err != nil {
				mu.Lock()
				result["inactive_bounces_error"] = err.Error()
				mu.Unlock()
				return
			}
			mu.Lock()
			result["inactive_bounces"] = jsonLen(extractList(data, "Bounces"))
			mu.Unlock()
		}()

		// Configured message streams.
		wg.Add(1)
		go func() {
			defer wg.Done()
			data, err := serverGet(c, "message-streams", nil)
			if err != nil {
				mu.Lock()
				result["streams_error"] = err.Error()
				mu.Unlock()
				return
			}
			var parsed any
			json.Unmarshal(extractList(data, "MessageStreams"), &parsed)
			mu.Lock()
			result["streams"] = parsed
			mu.Unlock()
		}()

		wg.Wait()

		out, _ := json.Marshal(result)
		return printData("", out)
	},
}
