#!/usr/bin/env bash

set -e
set -o pipefail

version=v0.0.1
target_dir="/usr/local/bin"
[ -d "$target_dir" ] || target_dir="/usr/bin"
[ -d "$target_dir" ] || { echo "- Error: Could not find /usr/local/bin nor /usr/bin directories." >&2; exit 1; }

_sudo="sudo"
command -v sudo &> /dev/null || _sudo=""

url_override=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --help)
            _usage
            ;;
        --url)
            url_override="$2"
            shift # past argument
            shift # past value
            ;;
        *)
            echo "Unknown option: $1" >&2
            _usage
            ;;
    esac
done


function _usage() {
    echo "Usage: $(basename "$0") [--help]"
    echo "Install Bayesh." 
    echo "Options:"
    echo "  --help         Show this help message and exit"
    echo "  --url <url>    Overwrite the download URL"
    exit 0
}

function _check_dependency() {
    command -v "$1" &> /dev/null || { echo "- Error: Required dependency $1 is not installed." >&2; exit 1; }
}

function _install_bayesh(){
    arc=$1
    if [[ -n "$url_override" ]]; then
        url="$url_override"
    else
        url="https://github.com/mads-bisgaard/bayesh/releases/download/${version}/bayesh-${version}-linux-${arc}.tar.gz"
    fi
    echo "- downloading Bayesh ${version} for architecture ${arc} to ${target_dir}/bayesh"
    ${_sudo} curl -sSL "$url" | tar -xzf - -C "${target_dir}"
    ${_sudo} chmod +x "${target_dir}/bayesh"
    command -v "bayesh" &> /dev/null || { echo "- Error: bayesh could not be found after installation." >&2; exit 1; }
}

function _print_bayesh() {
    CYAN="\033[0;36m"
    RESET="\033[0m"

    # Hardcoded ASCII Art for "BAYESH"
    echo -e "${CYAN}"
    echo "░████████                                               ░██        "
    echo "░██    ░██                                              ░██        "
    echo "░██    ░██   ░██████   ░██    ░██  ░███████   ░███████  ░████████  "
    echo "░████████         ░██  ░██    ░██ ░██    ░██ ░██        ░██    ░██ "
    echo "░██     ░██  ░███████  ░██    ░██ ░█████████  ░███████  ░██    ░██ "
    echo "░██     ░██ ░██   ░██  ░██   ░███ ░██               ░██ ░██    ░██ "
    echo "░█████████   ░█████░██  ░█████░██  ░███████   ░███████  ░██    ░██ "
    echo "                              ░██                                  "
    echo "                        ░███████                                    "
    echo -e "${RESET}"
    echo "- For documentation, see https://github.com/mads-bisgaard/bayesh"    
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
        _install_bayesh "amd64"
        ;;
    armv7l|armv6l|arm)
        _install_bayesh "arm"
        ;;
    *)
        echo "Unsupported architecture: $arch. Please file an issue on https://github.com/mads-bisgaard/bayesh and I will add support for your architecture."
        exit 1
        ;;
esac

_print_bayesh