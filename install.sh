#!/usr/bin/env bash

set -e
set -o pipefail

version=v0.0.1
target_dir="/usr/local/bin"
[ -d "$target_dir" ] || target_dir="/usr/bin"
[ -d "$target_dir" ] || { echo "- Error: Could not find /usr/local/bin nor /usr/bin directories." >&2; exit 1; }

_sudo="sudo"
command -v sudo &> /dev/null || _sudo=""

function _usage() {
    echo "Usage: $(basename "$0") [--help]"
    echo "Install Bayesh." 
    echo "Options:"
    echo "  --help   Show this help message and exit"
    exit 0
}

function _check_dependency() {
    command -v "$1" &> /dev/null || { echo "- Error: Required dependency $1 is not installed." >&2; exit 1; }
}

function _install_bayesh(){
    arc=$1
    url="https://github.com/mads-bisgaard/bayesh/releases/download/${version}/bayesh-${version}-linux-${arc}.tar.gz"
    echo "- downloading Bayesh ${version} for architecture ${arc} to ${target_dir}/bayesh"
    ${_sudo} curl -sSL "$url" | tar -xzf - -C "${target_dir}"
    ${_sudo} chmod +x "${target_dir}/bayesh"
    command -v "bayesh" &> /dev/null || { echo "- Error: bayesh could not be found after installation." >&2; exit 1; }
}

while [[ $# -gt 0 ]]; do
    case $1 in
        --help)
            _usage
            ;;
        *)
            echo "Unknown option: $1" >&2
            _usage
            ;;
    esac
done


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
        _install_bayesh "amd64"
        ;;
    *)
        echo "Unsupported architecture: $arch. Please file an issue on https://github.com/mads-bisgaard/bayesh and I will add support for your architecture."
        exit 1
        ;;
esac

echo "- done installing Bayesh"
echo "- set up your shell integration to get the most out of Bayesh"
echo "- for documentation, see https://github.com/mads-bisgaard/bayesh"