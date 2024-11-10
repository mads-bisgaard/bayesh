#!/bin/bash

REPO_DIR=$(realpath "$(dirname "${BASH_SOURCE[0]}")")
BAYESH_DIR="${REPO_DIR}/tmp"
export BAYESH_DIR

fzf --help > /dev/null
xargs --help > /dev/null

# shellcheck disable=SC2139
alias bayesh="${REPO_DIR}/.venv/bin/python3 -m bayesh"
bayesh --help > /dev/null

source "${REPO_DIR}/shell/bayesh.bash"

