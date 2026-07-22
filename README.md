# tmuxssel

A tmux session picker built on [fzf](https://github.com/junegunn/fzf), plus a
per-window "flag" that long-running tasks can raise. The picker sorts sessions
containing flagged windows to the top, and the tmux status line can mark each
flagged window individually.

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
(window flag management).

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

Sessions containing a flagged window are listed first, each marked with a
leading `●`.

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

Each window carries a boolean flag, stored as the `@ssel_flag` tmux user
option on the window itself, so it lives and dies with the window — there is
no external state to keep in sync. A session containing a flagged window
sorts to the top of the picker. The intended use is for a long-running task
to raise its window's flag when it finishes, and for the status bar to mark
the windows that are waiting.

```
tmuxssel flag {get|set|clear|toggle|count} [WINDOW]
```

- `get` — print `1` if the window is flagged, otherwise `0`.
- `set` / `clear` — raise or lower the flag.
- `toggle` — flip the flag.
- `count` — print the number of flagged windows.

`WINDOW` is any tmux window target (a window ID such as `@5`, or a name such
as `mysession:mywindow`) and defaults to the current window (resolved from
`$TMUX_PANE`), so a task running inside its own window can raise its flag on
completion with:

```sh
make deploy; tmuxssel flag set
```

Bind toggling the current window's flag to a key. `run-shell` does not set
`$TMUX_PANE`, so pass the window in explicitly — tmux expands `#{window_id}`
to the window the key was pressed in:

```
bind-key F run-shell "tmuxssel flag toggle '#{window_id}'"
```

## Status bar

Because the flag is a window option, the status line's window list can show
it directly: `window-status-format` is expanded once per window, so a
`#{?...}` conditional on `@ssel_flag` marks exactly the flagged windows, with
no external command involved:

```
set -g window-status-format '#{?#{@ssel_flag},#[fg=red]● #[default],}#I:#W#F'
set -g window-status-current-format '#{?#{@ssel_flag},#[fg=red]● #[default],}#I:#W#F'
```

`count` is a single global number, so it can be called from a `#(...)`
command:

```
set -g status-right "#(tmuxssel flag count) flagged"
```

tmux runs `#(...)` commands asynchronously, so after a flag changes the count
updates a moment later. `flag` mutations call `refresh-client -S` to trigger
that redraw promptly.

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
