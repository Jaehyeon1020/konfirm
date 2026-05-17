# konfirm status/help UX design

## Summary

Improve konfirm's day-to-day CLI usability without adding new commands or broad setup automation.

The change focuses on two existing surfaces:

- `konfirm status`: show a compact diagnostic view of the current context, config, allowlist, dependencies, and shell setup hints.
- `konfirm -h`, `konfirm --help`, `konfirm help`: provide a conventional CLI help page with clearer commands, options, and examples.

No `--json` output is included. The status output remains human-readable only.

## Goals

- Make `konfirm status` useful for team onboarding and support: a teammate can run it and share the output.
- Keep `konfirm status` compact enough for regular terminal use.
- Make help output feel like a normal CLI help page rather than a minimal usage stub.
- Avoid turning help into a full installation or setup guide.
- Preserve existing command names and core behavior.

## Non-goals

- Add a new `doctor` or `setup` command.
- Add `konfirm status --json`.
- Automatically modify shell configuration files.
- Add team policy import/export or shared config distribution.
- Redesign the interactive `konfirm config` fzf flow.

## User-facing design

### `konfirm status`

`konfirm status` continues to accept no arguments. Passing arguments remains a usage error.

The output becomes section-based:

```text
Context
  current: prod-cluster

Config
  path: /Users/me/Library/Application Support/konfirm/config.json
  allowed for current context:
    get
    logs
    describe

Dependencies
  kubectl: found
  fzf: found

Shell setup
  detected shell: zsh
  completion: add `source <(konfirm completion zsh)` to ~/.zshrc
  recommended alias: alias k="konfirm kubectl"
```

The exact path follows `store.ConfigPath()`, so it remains platform-specific.

If the current context has no allowed subcommands, the allowlist line shows `(none)`.

If the current context cannot be resolved, status still prints other available diagnostics:

```text
Context
  current: unavailable (<error>)
```

When the current context is unavailable, the allowlist section should state that it cannot be shown for the current context.

Dependency checks are informational:

- `kubectl` is `found` or `missing`.
- `fzf` is `found` or `missing`.

Shell detection uses the `SHELL` environment variable basename. Unknown or empty shells are shown as `unknown`.

Completion detection is intentionally conservative. The command does not attempt to parse shell rc files deeply. It prints shell-specific setup guidance when the shell is recognized:

- zsh: `source <(konfirm completion zsh)`
- fish: `konfirm completion fish | source`
- unknown shell: point users to `konfirm completion zsh` and `konfirm completion fish`

Alias detection is not attempted. The output shows a recommended alias:

```text
recommended alias: alias k="konfirm kubectl"
```

### `konfirm -h`

Help output should stay concise and conventional.

It should include:

- Short description of what konfirm does.
- Usage section.
- Commands section with one-line descriptions.
- Options section with `-h, --help`.
- Examples section with common commands.
- A final short pointer to `konfirm status`.

The help page should not include long shell setup snippets or a full onboarding guide.

Suggested shape:

```text
konfirm - confirm the effective kubectl context before execution

Usage:
  konfirm kubectl <kubectl args...>
  konfirm <command> [args...]

Commands:
  kubectl      Run kubectl after confirming the effective context
  config       Interactively manage allowed kubectl subcommands
  add          Allow kubectl subcommands for the current context
  remove       Remove allowed kubectl subcommands for the current context
  status       Show current context, allowlist, and local setup
  completion   Generate shell completion script
  version      Show konfirm version

Options:
  -h, --help   Show help

Examples:
  konfirm kubectl get pods
  konfirm kubectl --context prod-cluster get deploy
  konfirm add get logs
  konfirm config
  konfirm status

Run `konfirm status` to inspect your current context, allowlist, and local setup.
```

## Internal design

### Status package

Keep the public command entrypoint as `status.Run(args []string) int`.

Internally, split the status flow into small units:

- collect current context
- collect config path and config contents
- collect dependency availability
- collect shell information
- render the status report

The status command may continue when some diagnostic collection fails:

- Current context lookup failure does not stop the command.
- Missing `kubectl` or `fzf` does not stop the command.
- Unknown shell does not stop the command.

Config path resolution or config load failure is more significant:

- If config path resolution fails, show the error in the Config section and return exit code `1`.
- If config JSON cannot be read or parsed, show the error in the Config section and return exit code `1`.
- A missing config file is not an error; it means no allowlist is configured.

### Support package

Keep help rendering in `internal/commands/support`.

`support.Usage` becomes the single source for help text used by:

- no command
- `-h`
- `--help`
- `help`
- unknown command usage follow-up

The ASCII logo may remain, but the help text after it should be concrete and scannable.

`support.Usage(w)` should write the full help output to the provided writer, including the logo if the logo remains.

## Testing plan

Add focused tests because these changes are output-sensitive.

### Help tests

Add tests around `support.Usage` that assert the output includes:

- description
- Usage
- Commands
- Options
- Examples
- representative commands such as `kubectl`, `config`, `status`

### Status tests

Avoid tests that require a real Kubernetes cluster or installed `kubectl`.

Introduce small dependency-injected helpers or render functions so tests can cover:

- context available with allowed subcommands
- context available with no allowed subcommands
- context unavailable
- missing dependencies
- zsh, fish, and unknown shell guidance
- config load error path

### Verification

Run:

```bash
go test ./...
go build ./...
```

## Acceptance criteria

- `konfirm status` shows context, config path, allowlist, dependency status, and shell setup guidance.
- `konfirm status` remains optionless; no `--json` support is added.
- `konfirm status` degrades gracefully when context, dependencies, or shell detection fail.
- `konfirm -h`, `konfirm --help`, and `konfirm help` show a clearer conventional CLI help page.
- Help output remains concise and does not become an installation guide.
- Tests cover the new help and status rendering behavior.
