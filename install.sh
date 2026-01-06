#!/usr/bin/env sh

set -e

version=v0.0.2
_sudo="sudo"
command -v sudo > /dev/null 2>&1 || _sudo=""
target_dir="/usr/local/bin"
[ -d "$target_dir" ] || target_dir="/usr/bin"
[ -d "$target_dir" ] || { echo "- Error: Could not find /usr/local/bin nor /usr/bin directories." >&2; exit 1; }

arch=$(uname -m)
case "$arch" in
    x86_64)
        goarch="amd64"
        ;;
    armv7l|armv6l|arm)
        goarch="arm"
        ;;
    *)
        echo "Unsupported architecture: $arch. Please file an issue on https://github.com/mads-bisgaard/bayesh and I will add support for your architecture."
        exit 1
        ;;
esac

_usage() {
    echo "Usage: install.sh [--help] [--url <url>]"
    echo "Install Bayesh." 
    echo "Options:"
    echo "  --help         Show this help message and exit"
    echo "  --url <url>    Overwrite the download URL"
    exit 0
}

url="https://github.com/mads-bisgaard/bayesh/releases/download/${version}/bayesh-${version}-linux-${goarch}.tar.gz"
while [ $# -gt 0 ]; do
    case $1 in
        --help)
            _usage
            ;;
        --url)
            url="$2"
            shift
            shift
            ;;
        *)
            echo "Unknown option: $1" >&2
            _usage
            ;;
    esac
done


_check_dependency() {
    command -v "$1" > /dev/null 2>&1 || { echo "- Error: Required dependency $1 is not installed." >&2; exit 1; }
}

_install_bayesh(){
    echo "- downloading Bayesh ${version} for architecture ${goarch} to ${target_dir}/bayesh"
    ${_sudo} curl -sSL "$url" | ${_sudo} tar -xzf - -C "${target_dir}"
    ${_sudo} chmod +x "${target_dir}/bayesh"
    command -v "bayesh" > /dev/null 2>&1 || { echo "- Error: bayesh could not be found after installation." >&2; exit 1; }
}

_print_bayesh() {
    CYAN="\033[94m"
    RESET="\033[0m"

    printf "%b\n" "${CYAN}"
    echo "░████████                                               ░██        "
    echo "░██    ░██                                              ░██        "
    echo "░██    ░██   ░██████   ░██    ░██  ░███████   ░███████  ░████████  "
    echo "░████████         ░██  ░██    ░██ ░██    ░██ ░██        ░██    ░██ "
    echo "░██     ░██  ░███████  ░██    ░██ ░█████████  ░███████  ░██    ░██ "
    echo "░██     ░██ ░██   ░██  ░██   ░███ ░██               ░██ ░██    ░██ "
    echo "░█████████   ░█████░██  ░█████░██  ░███████   ░███████  ░██    ░██ "
    echo "                              ░██                                  "
    echo "                        ░███████                                    "
    printf "%b\n" "${RESET}"
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
_install_bayesh
_print_bayesh