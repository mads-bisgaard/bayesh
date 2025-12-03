#!/usr/bin/env bash

set -e
set -o pipefail

version=v0.0.1

function usage() {
    echo "Usage: $(basename "$0") <shell> [-y]"
    echo "Install Bayesh. Supported shells: bash, zsh."
    exit 1
}

function _check_dependency() {
    command -v "$1" &> /dev/null || { echo "- Error: Required dependency $1 is not installed." >&2; exit 1; }
}

function _download_bayesh(){
    arc=$1
    url="https://github.com/repos/mads-bisgaard/bayesh/releases/download/v$version/bayesh-${version}-linux-${arc}"
    sudo curl -sSL "${url}" -o /usr/local/bin/bayesh
    sudo chmod +x /usr/local/bin/bayesh
    command -v "bayesh" &> /dev/null || { echo "- Error: bayesh could not be found after installation." >&2; exit 1; }
}

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
_check_dependency "tar"

arch=$(uname -m)
case "$arch" in
    x86_64)
        _download_bayesh "amd64"
        ;;
    *)
        echo "Unsupported architecture: $arch. Please file an issue on https://github.com/mads-bisgaard/bayesh and I will add support for your architecture."
        exit 1
        ;;
esac

echo "- done installing Bayesh"
echo "- restart your terminal and open Bayesh by using Ctrl-e"
echo "- for documentation, see https://github.com/mads-bisgaard/bayesh"