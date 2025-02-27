#!/usr/bin/env bash

set -uo pipefail

REPO_DIR=$(realpath "$(dirname "${BASH_SOURCE[0]}")")

function _check_exists() {
    [[ -e "$1" ]] || { echo "- $1 does not exist"; exit 1; }
}

function _check_dependency() {
    command -v "$1" &> /dev/null || { echo "- Error: Required dependency $1 is not installed." >&2; exit 1; }
}

function allow() {
  while true; do
    read -rp "$1 ([y]/n) " answer
    if [[ $answer = y ]]; then
      return 0
    elif [[ $answer = n ]]; then
      return 1
    fi
  done
}

echo "- checking dependencies are installed"
_check_dependency "python3"
_check_dependency "fzf"
_check_dependency "awk"
_check_dependency "md5sum"
_check_dependency "cut"
_check_dependency "head"
_check_dependency "tail"
_check_dependency "echo"
_check_dependency "grep"
_check_dependency "curl"
_check_dependency "jq"

echo "- setting up python venv"
python3 -m venv "${REPO_DIR}/.venv"
_check_exists "${REPO_DIR}/.venv/bin/python3"
echo "- installing bayesh cli"
"${REPO_DIR}/.venv/bin/python3" -m pip install "${REPO_DIR}" &> /dev/null
_check_exists "${REPO_DIR}/.venv/bin/bayesh"
echo "- adding bayesh executable to bin directory"
[[ -e "${REPO_DIR}/bin/bayesh" ]] && rm "${REPO_DIR}/bin/bayesh"
ln -s "${REPO_DIR}/.venv/bin/bayesh" "${REPO_DIR}/bin/bayesh"

_shell=$(basename "$SHELL")
[[ "$_shell" = "bash" ]] || [[ "$_shell" = "zsh" ]] || { echo "Currently Bayesh is only compatible with zsh and bash" >&2; exit 1; }
_rcfile="$HOME/.${_shell}rc"

if allow "- Add Bayesh to PATH (required for Bayesh to be functional)?"; then
    # shellcheck disable=SC2016
    echo 'export PATH="$PATH:'"${REPO_DIR}/bin"'"' >> "$_rcfile"
fi

if allow "- Add $_shell integration (required for Bayesh to be functional)?"; then
    echo "source ${REPO_DIR}/shell/bayesh.${_shell}" >> "$_rcfile"
fi

echo "- done installing Bayesh. See https://github.com/mads-bisgaard/bayesh for documentation"
exec "$_shell"