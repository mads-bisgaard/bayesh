# Agent instructions

Bayesh is a Go application with Zsh shell integration. One of the project's
most important priorities is the user experience of Bayesh's Zsh integration
inside tmux. Aim to make that experience as polished, predictable, and
frictionless as possible. Evaluate user-facing changes in the context of an
interactive Zsh session running inside tmux, and preserve consistency between
the implementation and its tests.

## Core development values

- Always prefer simple solutions.
- Use test driven development. 
- Keep changes focused and update tests when behavior changes.
- For implementations in go, follow design patterns and pilosophy of https://github.com/junegunn/fzf.
- When commiting work into git, commit it with your own credentials.

## Validation

Run the relevant checks from the repository root:

```sh
go test ./...
make build
PATH="$PWD/build:$PATH" bats --print-output-on-failure tests
```

## Manual testing

Test user-facing changes in a real tmux session. `make manual-test` builds
Bayesh and starts a detached tmux session with a fresh database, separate from
your own shell configuration and `~/.bayesh`. Suggestions need some history,
so run a few commands before opening Bayesh with `Ctrl-e`:

```sh
make manual-test
tmux -L bayesh-manual-test send-keys 'echo one' Enter 'echo two' Enter 'echo one' Enter
tmux -L bayesh-manual-test send-keys C-e                 # also C-Up, C-Down and C-Right
tmux -L bayesh-manual-test capture-pane -p -t bayesh.0   # the shell
tmux -L bayesh-manual-test capture-pane -p -t bayesh.1   # the fzf pane
make manual-test-stop
```

Give Bayesh a moment to respond before capturing a pane. To use the session
interactively, run `tmux -L bayesh-manual-test attach`.


# Documentation of software arcitechture and design
Under [bayesh-dev-notes](./bayesh-dev-notes) is an Obsidean vault with software design and architecture docs.
Make sure to keep this up to date as you develop. If it is not already there, create it. 
Keep design notes concise and precise.