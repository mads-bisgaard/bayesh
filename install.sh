#!/usr/bin/env bash

set -e
set -o pipefail

DIR="$HOME/.bayesh/bin"

# Function to display usage
usage() {
    echo "Usage: $(basename "$0") <shell> [-y]"
    echo "Install Bayesh. Supported shells: bash, zsh."
    exit 1
}

shell=$1
[[ "$shell" == "bash" || "$shell" == "zsh" ]] || usage
shift 

function _check_exists() {
    [[ -e "$1" ]] || { echo "- Error: Something unexpected happened. $1 does not exist"; exit 1; }
}

function _check_dependency() {
    command -v "$1" &> /dev/null || { echo "- Error: Required dependency $1 is not installed." >&2; exit 1; }
}

echo "- checking dependencies are installed"
_check_dependency "fzf"
_check_dependency "awk"
_check_dependency "md5sum"
_check_dependency "cut"
_check_dependency "head"
_check_dependency "tail"
_check_dependency "echo"
_check_dependency "grep"
_check_dependency "curl"
_check_dependency "tar"

echo "- creating installation directory"
mkdir -p "${DIR}"

echo "- detecting OS and architecture"
os_name=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)

case "$arch" in
    x86_64)
        arch="amd64"
        ;;
    *)
        echo "Unsupported architecture: $arch. Please file an issue on https://github.com/mads-bisgaard/bayesh and I will add support for your architecture."
        exit 1
        ;;
esac

echo "- downloading latest bayesh binary from github"
search_pattern="${os_name}-${arch}"
url=$(curl -s https://api.github.com/repos/mads-bisgaard/bayesh/releases/latest | grep "browser_download_url.*${search_pattern}.*\.tar\.gz" | sed -E 's/.*"browser_download_url": "(.*)".*/\1/')
curl -sSL "${url}" | tar -xz -C "${DIR}"
_check_exists "${DIR}/bayesh"

_rcfile="$HOME/.${shell}rc"

echo "- exporting PATH"
# shellcheck disable=SC2016
echo 'export PATH="$PATH:'"${DIR}"'"' >> "$_rcfile"
echo "- sourcing bayesh.${shell}"
echo "source ${DIR}/bayesh.${shell}" >> "$_rcfile"

echo "- done installing Bayesh"
echo "- restart your terminal and open Bayesh by using Ctrl-e"
echo "- for documentation, see https://github.com/mads-bisgaard/bayesh"