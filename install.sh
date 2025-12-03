#!/usr/bin/env bash

set -e
set -o pipefail

version=v0.0.1
target_path="/usr/local/bin/bayesh"

function _usage() {
    echo "Usage: $(basename "$0") [--help] [--from-directory <path>]"
    echo "Install Bayesh." 
    echo "Options:"
    echo "  --help                    Show this help message and exit"
    echo "  --from-directory <path>   Install Bayesh from local directory instead of downloading"
    exit 0
}

function _check_dependency() {
    command -v "$1" &> /dev/null || { echo "- Error: Required dependency $1 is not installed." >&2; exit 1; }
}

function _install_bayesh(){
    arc=$1
    from_directory=$2
    
    if [[ -n "$from_directory" ]]; then
        sudo cp "${from_directory}/bayesh-${version}-linux-${arc}" "${target_path}"
    else
        url="https://github.com/repos/mads-bisgaard/bayesh/releases/download/${version}/bayesh-${version}-linux-${arc}"
        sudo curl -sSL "${url}" -o "${target_path}"
    fi
    
    sudo chmod +x "${target_path}"
    command -v "bayesh" &> /dev/null || { echo "- Error: bayesh could not be found after installation." >&2; exit 1; }
}

from_directory=""
while [[ $# -gt 0 ]]; do
    case $1 in
        --help)
            _usage
            ;;
        --from-directory)
            from_directory="$2"
            shift 2
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
        _install_bayesh "amd64" "$from_directory"
        ;;
    *)
        echo "Unsupported architecture: $arch. Please file an issue on https://github.com/mads-bisgaard/bayesh and I will add support for your architecture."
        exit 1
        ;;
esac

echo "- done installing Bayesh"
echo "- set up your shell integration to get the most out of Bayesh"
echo "- for documentation, see https://github.com/mads-bisgaard/bayesh"