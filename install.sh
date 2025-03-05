#!/usr/bin/env bash

set -o pipefail

REPO_DIR=$(realpath "$(dirname "${BASH_SOURCE[0]}")")

# Function to display usage
usage() {
    echo "Usage: $(basename "$0") <shell> [-y]"
    echo "Install Bayesh. Supported shells: bash, zsh."
    echo "Add -y argument to automatically answer 'yes' for automatic confirmation."
    exit 1
}

shell=$1
[[ "$shell" == "bash" || "$shell" == "zsh" ]] || usage
shift 

# Default value for the confirmation flag
automatic_confirm=false

# Parse command line arguments
while getopts ":y" opt; do
    case ${opt} in
        y )
            automatic_confirm=true
            ;;
        \? )
            usage
            ;;
    esac
done

# Shift the parsed options away
shift $((OPTIND -1))

function _check_exists() {
    [[ -e "$1" ]] || { echo "- Error: Something unexpected happened. $1 does not exist"; exit 1; }
}

function _check_dependency() {
    command -v "$1" &> /dev/null || { echo "- Error: Required dependency $1 is not installed." >&2; exit 1; }
}

function allow() {
  while true; do
    echo "$1 ([y]/n) "
    read -r answer
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
rm -f "${REPO_DIR}/bin/bayesh" && ln -s "${REPO_DIR}/.venv/bin/bayesh" "${REPO_DIR}/bin/bayesh"

_rcfile="$HOME/.${shell}rc"

if "$automatic_confirm" || allow "Add Bayesh to PATH (required for Bayesh to be functional)?"; then
    echo "- exporting PATH"
    # shellcheck disable=SC2016
    echo 'export PATH="$PATH:'"${REPO_DIR}/bin"'"' >> "$_rcfile"
fi

if "$automatic_confirm" || allow "Add $shell integration (required for Bayesh to be functional)?"; then
    echo "- sourcing bayesh.${shell}"
    echo "source ${REPO_DIR}/shell/bayesh.${shell}" >> "$_rcfile"
fi

echo "- done installing Bayesh"
echo "- restart your terminal and open Bayesh by using Ctrl-e"
echo "- for documentation, see https://github.com/mads-bisgaard/bayesh"