# tmuxssel

A tmux session picker built on [fzf](https://github.com/junegunn/fzf), plus a
per-session "flag" that long-running tasks can raise and that the picker sorts
to the top.

The picker shows a fuzzy-searchable menu containing:

- existing tmux sessions, and
- git repositories found by walking the filesystem (any directory containing
  a `.git` entry, including the `.git` files used by worktrees).

Selecting an entry switches to the corresponding session, creating it first
if it doesn't already exist. New sessions start in the directory of the
repository they were created from.

Candidates are streamed into fzf as they're discovered, so the menu is
usable immediately even while a large filesystem walk is still in progress.

## Requirements

- Python 3 (standard library only)
- tmux
- fzf

## Installation

Copy or symlink the `tmuxssel` script onto your `PATH`:

```sh
ln -s "$PWD/tmuxssel" ~/bin/tmuxssel
```

## Usage

tmuxssel has two subcommands: `pick` (the session picker) and `flag`
(session flag management).

### `tmuxssel pick`

```
tmuxssel pick [--walk-start PATH] [--ignore FRAGMENT]
              [--find-and-replace FIND:REPLACE]
```

- `--walk-start PATH` — where to start the filesystem walk. Defaults to
  `$HOME`.
- `--ignore FRAGMENT` — skip directories whose path contains `FRAGMENT`. May
  be given multiple times.
- `--find-and-replace FIND:REPLACE` — rewrite repository paths into session
  names by replacing `FIND` with `REPLACE`. Applied in the order given. May
  be given multiple times.

Flagged sessions are listed first, each marked with a leading `●`.

#### Example

```sh
tmuxssel pick \
    --ignore .cache \
    --ignore go/pkg \
    --find-and-replace "$HOME/repos/github.com/:" \
    --find-and-replace ".:,"
```

With these directives, the repository at
`$HOME/repos/github.com/alice/website.com` appears in the menu as
`alice/website,com`.

To bind it to a tmux key (tmux 3.2+):

```
bind-key S display-popup -E "tmuxssel pick --ignore .cache ..."
```

### `tmuxssel flag`

Each session carries a boolean flag, stored as a tmux user option on the
session itself, so it lives and dies with the session — there is no external
state to keep in sync. Flagged sessions sort to the top of the picker. The
intended use is for a long-running task to raise its session's flag when it
finishes, and for the status bar to show how many sessions are waiting.

```
tmuxssel flag {get|set|clear|toggle|count} [SESSION]
```

- `get` — print `1` if the session is flagged, otherwise `0`.
- `set` / `clear` — raise or lower the flag.
- `toggle` — flip the flag.
- `count` — print the number of flagged sessions.

`SESSION` defaults to the current session (resolved from `$TMUX_PANE`), so a
task running inside its own session can raise its flag on completion with:

```sh
make deploy; tmuxssel flag set
```

Bind toggling the current session's flag to a key. `run-shell` does not set
`$TMUX_PANE`, so pass the session in explicitly — tmux expands
`#{session_name}` to the session the key was pressed in:

```
bind-key F run-shell "tmuxssel flag toggle #{session_name}"
```

## Status bar

`count` and `get` are meant to be wired into the tmux status bar. `count` is
a single global number, so it can be called directly:

```
set -g status-right "#(tmuxssel flag count) flagged"
```

`get` reports one specific session, so the status format must pass the session
name into the command. tmux caches each distinct `#(...)` command string
separately, so interpolating `#{session_name}` gives each session its own
result:

```
set -g status-right "#(tmuxssel flag get #{session_name}) #(tmuxssel flag count)"
```

`get` prints `1`/`0`; turning that into a marker is left to your status
configuration, because a tmux `#{?...}` conditional can't test the output of
an asynchronous `#(...)` command. A status-generator script can branch on it
directly.

tmux runs `#(...)` commands asynchronously, so after a flag changes the status
bar updates a moment later. `flag` mutations call `refresh-client -S` to
trigger that redraw promptly.

## Behaviour

- Existing sessions appear under their own names, unmodified. If a rewritten
  repository path collides with an existing session name, only the session
  is shown.
- tmux silently replaces `.` and `:` in session names with `_`, so a session
  created from a candidate containing those characters can't then be found
  by its menu name. Map them away with a directive such as
  `--find-and-replace ".:,"`.
- Inside tmux, selection uses `switch-client`; outside tmux, it uses
  `attach-session`.
- Session names are matched exactly (`has-session -t =NAME`), so `foo` is
  never confused with `foobar`.
- Cancelling fzf (Escape or Ctrl-C) exits without doing anything.
- Directories that can't be read during the walk are silently skipped.

## License

MIT — see [LICENSE](LICENSE).
