#!/bin/bash

set -euo pipefail

REPO_DIR=$(realpath "$(dirname "${BASH_SOURCE[0]}")")
BASH_RC=$(realpath ~/.bashrc)

function _check_exists() {
    local file_or_dir
    file_or_dir=$1
    if [[ ! -e "$file_or_dir" ]]; then
        echo "- $file_or_dir does not exist"
        exit 1
    fi
}

function _check_dependency() {
    command -v "$1" &> /dev/null || { echo "- Error: $1 is not installed." >&2; exit 1; }
}


echo "- checking dependencies are installed"
_check_dependency "python3"
_check_dependency "fzf"
_check_dependency "awk"
_check_dependency "md5sum"

echo "- setting up python venv"
python3 -m venv "${REPO_DIR}/.venv"
_check_exists "${REPO_DIR}/.venv/bin/python3"
echo "- installing bayesh into python venv"
"${REPO_DIR}/.venv/bin/python3" -m pip install "${REPO_DIR}" &> /dev/null
_check_exists "${REPO_DIR}/.venv/bin/bayesh"
_check_exists "/usr/local/bin"
echo "- exposing bayesh executable on PATH"
sudo ln -s "${REPO_DIR}/.venv/bin/bayesh" "/usr/local/bin/bayesh"
_check_dependency "bayesh"

echo "- done installing bayesh"
_check_exists "${BASH_RC}"
echo "- Add 'source ${REPO_DIR}/shell/setup.bash' to ${BASH_RC} and source ${BASH_RC} to activate bayesh"