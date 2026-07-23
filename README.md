# tmuxssel

A tmux session picker built on [fzf](https://github.com/junegunn/fzf). The
picker sorts sessions containing flagged windows to the top (see
[Window flags](#window-flags)).

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

```
tmuxssel [--walk-start PATH] [--ignore FRAGMENT]
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

### Example

```sh
tmuxssel \
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
bind-key S display-popup -E "tmuxssel --ignore .cache ..."
```

## Window flags

tmuxssel treats a window as flagged when the window's `@ssel_flag` tmux user
option is set. Because the flag is a window option, it lives and dies with
the window — there is no external state to keep in sync.

tmuxssel only reads the option; raising and clearing flags is left to tmux
key bindings and scripts:

```sh
tmux set-option -w @ssel_flag 1   # raise the current window's flag
tmux set-option -wu @ssel_flag    # clear it
```

The intended use is for a long-running task to raise its window's flag when
it finishes:

```sh
make deploy; tmux set-option -w @ssel_flag 1
```

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
