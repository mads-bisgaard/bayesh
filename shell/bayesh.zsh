#!/usr/bin/env zsh

# shellcheck source=shell/bayesh.sh
source "$(dirname $0)"/bayesh.sh

function bayesh_infer_cmd() {
    (_bayesh_infer_cmd)
}

add-zsh-hook precmd _bayesh_update
export BAYESH_AVOID_IF_EMPTY
