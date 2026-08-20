---
name: ggg
description: Operate the ggg CLI to clone, sync, inspect, and navigate git repositories declared in a YAML config file. Use when an AI agent is asked to install, configure, inspect, automate, troubleshoot, or run ggg, including bulk clone, pull, push, status, or import operations across many repositories.
---

# Operate GGG

Use the installed `ggg` executable as the source of truth for the available commands and flags.

## Preflight

1. Run `command -v ggg` and `ggg --version` before operating it.
2. Run `ggg <command> --help` before using flags whose behavior is unclear.
3. GGG reads `~/.config/ggg/repositories.yaml`. If it is missing, run `ggg init` and tell the user to declare repositories before any bulk operation.
4. Run `ggg doctor --json` when a command fails unexpectedly. It reports the config file, config syntax, base directory, duplicate URLs, clone status, and remote reachability.
5. Refresh this skill after upgrading GGG with `ggg skills install`. If GGG reports locally modified skill files, do not add `--force` without the user's approval.

## Inspect before mutating

These commands only read: `list`, `status`, `outdated`, `diff`, `doctor`, `cd`, `shell-init`.

Enumerate the exact repositories a command will touch with `ggg list --json` before running anything that writes. Add the same `--group` or `--filter` you intend to use so the preview matches the real selection.

## Scope every bulk operation

`clone`, `pull`, `push`, `stash`, `checkout`, and `diff` act on **every configured repository** when no target is given, in parallel goroutines.

- Always narrow with a repository name, `--filter/-f`, or `--group/-g` unless the user explicitly asked for all repositories.
- State the resolved repository list and the operation before running it.
- `push` publishes commits to remotes. `checkout` and `stash` move or hide uncommitted work. Confirm these with the user whenever the selection is wider than a single repository.
- `pull` skips dirty repositories by design. Do not work around that by stashing first unless the user asks.
- Without `--json` these commands prompt for confirmation; with `--json` they auto-confirm. Never add `--json` merely to bypass a prompt.

## Commands that change the configuration file

`init`, `add`, `remove`, and `import` write `~/.config/ggg/repositories.yaml`. `validate` can delete duplicate entries. `export` writes a copy of the config to a path you choose.

- Show the user the entry you are about to add or remove before running the command.
- `import` requires an authenticated `gh` CLI. With `--json` it requires an explicit account **and** repository argument and imports only that one repository; never treat JSON mode as "select all".

## Produce automation-friendly output

Use `--json` whenever the output is parsed. Parse standard output only; never parse the styled tables, prompts, or spinners.

- `config`, `open`, and `browse` launch an editor or a browser and do not support `--json`. They return a JSON error object with a non-zero exit code. Do not run them in non-interactive contexts.
- Bulk commands report per-repository failures in a non-empty `error` field on each item and still exit 0. Inspect those fields instead of relying on the exit code.
- `ggg cd <name>` prints a path; it cannot change the caller's directory. `ggg shell-init <shell>` defines a `ggg()` shell wrapper that turns `ggg cd` into a real chdir. In a script, use `cd "$(ggg cd <name>)"` instead.
