#!/bin/bash

BAYESH_DIR="${BAYESH_DIR:-~/.bayesh}"
export BAYESH_DIR

REPO_DIR=$(realpath "$(dirname "${BASH_SOURCE[0]}")/..")
# shellcheck disable=SC2139
alias bayesh="${REPO_DIR}/.venv/bin/python3 -m bayesh"
bayesh --help > /dev/null
fzf --help > /dev/null
xargs --help > /dev/null

# shellcheck source=shell/bayesh.bash
source "$(dirname "${BASH_SOURCE[0]}")/bayesh.bash"

if [[ -n "$PROMPT_COMMAND" ]]; then
    PROMPT_COMMAND="$PROMPT_COMMAND; bayesh_update"
else
    PROMPT_COMMAND='bayesh_update'
fi