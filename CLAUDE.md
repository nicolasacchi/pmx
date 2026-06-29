# CLAUDE.md — pmx

Go CLI for **Postmark**. Single binary, dual-audience output (human tables /
LLM JSON), built on [`clicore`](https://github.com/nicolasacchi/clicore). Fleet
sibling of `ddx`/`otx`/`gx`/`stx` — same UX, exit codes, agent-mode, write gate.

**API**: Postmark REST API. Base URL `https://api.postmarkapp.com`.

## Authentication — two scopes

Resolution order per scope (first non-empty wins):

1. `--server-token` / `--account-token` flags
2. `POSTMARK_SERVER_TOKEN` / `POSTMARK_ACCOUNT_TOKEN` env vars
3. `~/.config/pmx/config.toml` (project from `--project`, then `default_project`)

| Scope | Header | Endpoints |
|-------|--------|-----------|
| **server** | `X-Postmark-Server-Token` | messages, bounces, stats, suppressions, email, templates, streams, server, webhooks, inbound-rules, overview, otp |
| **account** | `X-Postmark-Account-Token` | servers, signatures, domains, data-removals |

A command only needs the scope it calls; the client raises a clear auth error
(exit 2) at request time if the needed token is unset. `pmx config doctor`
probes both scopes.

> SMTP credentials (used by the Rails app to *send*) are NOT API tokens. Mint
> API tokens in the Postmark UI: Server → API Tokens (server scope), Account →
> API Tokens (account scope).

## Headline use case — OTP tracing

OTP mail is sent over SMTP but Postmark tracks it identically, so the Messages
API gives full delivery visibility. `pmx otp <email>` fans out (messages +
latest event timeline + bounces) in one call. Caveat: `--tag otp` only works if
the sender sets an `X-PM-Tag` header (the OTP mailer may not) — fall back to
`--recipient` + `--subject`.

## Output

- **TTY** → go-pretty table (commands with a column def in `internal/output/table.go`); JSON otherwise.
- **Pipe / `--json` / `--jq`** → JSON. `--jq` is gjson, NOT jq (`#.field`, `#.{a:A}`).
- **`CLAUDECODE=1`** → JSON + 100-row cap with a `note:` on stderr (`--limit 0` for all).
- List commands unwrap Postmark's `{TotalCount, <Key>:[...]}` envelope to the bare array.

## Exit codes (fleet-canonical, `clicore/cierrors`)

`0` ok · `1` generic/network · `2` auth (401/403/missing token) · `3` validation
(Postmark **422**/bad args) · `4` not found · `5` rate limited · `6` write refused
(`write_locked`, needs `--yes`).

## Write safety

State-changing verbs refuse without `--yes`/`--confirm` (exit 6); `--dry-run`
previews. Reads are ungated. Gate every new mutation with `gateWrite(action)`.

## Architecture

```
cmd/pmx/main.go              Entry point, version, exit-code dispatch
internal/
  client/{client,errors}.go   Postmark HTTP client (dual-scope auth, retries), APIError
  commands/root.go            Root cmd, global flags, getClient/printData/fromDate
  commands/messages.go        EXEMPLAR group — mirror it for new groups
  commands/otp.go             Composite OTP-trace command
  commands/<group>.go         One file per command group (init() self-registers)
  commands/{config,confirm,helpers}.go
  config/config.go            TOML config (server/account tokens), flag>env>file
  output/{output,table,filter}.go   JSON/table/gjson dispatcher
  timeparse/timeparse.go      --from/--to → Postmark YYYY-MM-DD
```

## Conventions for new command groups

1. One file per group, package `commands`; register via `init()` (no root.go edits).
2. Prefix every package-level identifier with the group name (avoid collisions).
3. Reads: `serverGet`/`accountGet` (or `c.Get(ctx(), client.Scope*, …)`).
   Writes: `c.Post/Put/Patch/Delete` after `gateWrite(...)`.
4. List → `printData("group.list", extractList(data, "<Key>"))` + a column def in `table.go`.
5. Single object / write result → `printData("", data)`.
6. `--file` bodies via `readJSONFile`; pagination via `pageParams()`; dates via `addDateParams()`.

## Commit checklist

1. `gofmt -w .`
2. `go build ./... && go vet ./...`
3. `go test -race ./...`
4. `make lint` (advisory until pre-existing errcheck findings are cleared)
