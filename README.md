# pmx — Postmark Explorer CLI

A single-binary Go CLI for the [Postmark](https://postmarkapp.com) API, built to
be useful for **humans** (pretty tables, sensible defaults, rich `--help`) **and
for LLMs / Claude Code** (clean JSON, agent row-caps, structured errors with
machine-readable exit codes). Fleet sibling of `ddx` / `otx` / `gx` / `stx`,
built on [`clicore`](https://github.com/nicolasacchi/clicore).

The headline use case: **trace a passwordless sign-in OTP** a customer says never
arrived. OTP mail is sent over SMTP, but Postmark tracks it identically, so the
Messages API still gives full delivery visibility.

```bash
pmx otp customer@example.com
```

## Install

```bash
go install github.com/nicolasacchi/pmx/cmd/pmx@latest
# or, from a clone:
make install
```

## Authentication

Postmark has **two token scopes**. pmx resolves each from **flag → env → config**
(first non-empty wins); a command only needs the scope it actually calls.

| Scope | Header | Flag / env | Used by |
|-------|--------|------------|---------|
| Server | `X-Postmark-Server-Token` | `--server-token` / `POSTMARK_SERVER_TOKEN` | messages, bounces, stats, suppressions, email, templates, streams, server, webhooks, inbound-rules, overview, otp |
| Account | `X-Postmark-Account-Token` | `--account-token` / `POSTMARK_ACCOUNT_TOKEN` | servers, signatures, domains, data-removals |

> **Tokens ≠ SMTP credentials.** The Rails app sends via Postmark *SMTP*
> username/password; the API tokens here are separate (Postmark UI → Servers →
> *API Tokens*, and Account → *API Tokens*).

### Multi-project config (`~/.config/pmx/config.toml`)

```bash
pmx config add production --server-token TOKEN --account-token TOKEN
pmx config add staging    --server-token TOKEN
pmx config use production
pmx config doctor          # probes each scope and reports what it can reach
```

```toml
default_project = "production"

[projects.production]
server_token  = "..."
account_token = "..."
```

## Quickstart — the OTP lens

```bash
# Everything about a recipient's recent mail in one shot:
pmx otp customer@example.com --from 7d

# Or step by step:
pmx messages outbound --recipient customer@example.com --from 2d
pmx messages outbound events <message-id>     # delivery / open / bounce timeline
pmx messages outbound dump   <message-id>     # raw rendered body

# Is the address hard-bounced / suppressed (silently dropping future OTPs)?
pmx bounces list --inactive --email customer@example.com
pmx suppressions list --stream outbound --jq '#.EmailAddress'
pmx bounces activate <id> --yes               # reactivate after the user fixes their inbox
```

## Command reference

### Server scope

```bash
pmx messages outbound [--recipient --from-email --tag --status --stream --subject --from --to]
pmx messages outbound get|dump|events <messageid>
pmx messages inbound  [--recipient --from-email --subject]
pmx messages inbound  get|dump <messageid>

pmx bounces list [--type --email --tag --message-id --inactive --from --to]
pmx bounces get|dump <id>
pmx bounces activate <id> --yes
pmx bounces tags
pmx bounces deliverystats

pmx stats outbound|sends|bounces|spam|opens|clicks [--from --to --tag]

pmx suppressions list   --stream outbound
pmx suppressions create --email x@y.com --stream outbound --yes
pmx suppressions delete --email x@y.com --stream outbound --yes   # reactivates

pmx email send --from … --to … --subject … --text … [--html --tag --stream] --yes
pmx email send-template --from … --to … (--template-id N | --alias A) --model-file m.json --yes
pmx email send-batch --file batch.json --yes

pmx templates list | get <id|alias> | validate --file t.json
pmx templates create --file t.json --yes | edit <id> --file t.json --yes | delete <id> --yes

pmx streams list | get <id>
pmx streams create --file s.json --yes | edit <id> --file s.json --yes | archive <id> --yes

pmx server get | edit --file s.json --yes

pmx webhooks list [--stream] | get <id>
pmx webhooks create --file w.json --yes | edit <id> --file w.json --yes | delete <id> --yes

pmx inbound-rules list
pmx inbound-rules create --rule spammer@bad.com --yes | delete <id> --yes

pmx overview      # parallel snapshot: delivery stats + inactive bounces + streams
```

### Account scope

```bash
pmx servers    list [--name] | get <id> | create --file | edit <id> --file | delete <id>   (--yes on writes)
pmx signatures list | get <id> | create/edit/delete (--file, --yes) | resend <id> | request-dkim <id>
pmx domains    list | get <id> | create/edit/delete (--file, --yes)
pmx domains    verify-dkim <id> | verify-returnpath <id> | rotate-dkim <id>   (--yes)
pmx data-removals create --file r.json --yes | status <id>
```

## Output

- **TTY** → a go-pretty table for commands with a column definition; JSON otherwise.
- **Piped / `--json` / `--jq`** → JSON.
- **`--jq <path>`** is [gjson](https://github.com/tidwall/gjson), *not* jq:
  arrays `#.field`, projections `#.{a:Field,b:Other}`, first element `0`.
- **Agent mode (`CLAUDECODE=1`)** → JSON, and list output is capped to 100 rows
  (shape-preserving) with a `note:` on stderr; pass `--limit 0` for all rows.

Postmark JSON keys are PascalCase (`MessageID`, `Recipients`, `Status`) — list
commands unwrap the `{TotalCount, <Key>: [...]}` envelope to the bare array so
`--jq` paths stay simple and the row-cap applies.

## Exit codes (fleet-canonical, `clicore/cierrors`)

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Generic / network |
| 2 | Auth (401/403, missing token) |
| 3 | Validation (Postmark 422 / bad args) |
| 4 | Not found (404) |
| 5 | Rate limited (429) |
| 6 | Write refused — needs `--yes` (`write_locked`) |

## Write safety

State-changing verbs (`email send`, `bounces activate`, `suppressions
create/delete`, all `*delete`/`edit`/`create`/`verify`/`rotate`/`archive`,
`data-removals create`) refuse unless `--yes`/`--confirm` is passed, exiting `6`;
`--dry-run` previews the intended mutation and sends nothing. Reads are ungated.

## Architecture

```
cmd/pmx/main.go              Entry point, version injection, exit-code dispatch
internal/
  client/client.go            HTTP client, dual-scope auth headers, retries
  client/errors.go            APIError + ExitCode() (clicore/cierrors), Postmark hints
  commands/root.go            Root command, global flags, getClient()
  commands/messages.go        EXEMPLAR group (others mirror it)
  commands/*.go               One file per command group
  commands/otp.go             Composite OTP-trace convenience command
  commands/config.go          Multi-project config + doctor
  config/config.go            TOML config (server/account tokens), resolution
  output/output.go            JSON/table dispatcher, TTY detection
  output/table.go             go-pretty tables, per-command column defs
  output/filter.go            gjson --jq filter
  timeparse/timeparse.go      --from/--to → Postmark YYYY-MM-DD
```

## Postmark gotchas baked in

- **OTP is SMTP-sent but API-visible** — `messages outbound` returns it. Tag
  filtering (`--tag otp`) only works if the sender sets an `X-PM-Tag` header,
  which the OTP mailer may not; fall back to `--recipient` + `--subject`.
- **Pagination** is `count` + `offset`, max 500/page; 429s are retried with
  `Retry-After`.
- **HTTP 422** (not 400) is Postmark's validation status → mapped to exit 3.

## License

MIT — see [LICENSE](LICENSE).
