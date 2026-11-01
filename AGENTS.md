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


## Validation

Run the relevant checks from the repository root:

```sh
go test ./...
make build
PATH="$PWD/build:$PATH" bats --print-output-on-failure tests
```
