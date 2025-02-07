#!/usr/bin/env zsh

# shellcheck source=shell/bayesh.sh
source "$(dirname $0)"/bayesh.sh

function bayesh_infer_cmd() {
    local cur
    local result
    local point
    
    cur="${CURSOR}"
    result=$(_bayesh_infer_cmd)
    line=$(echo "${result}" | tail -n 1);point=$(echo "${result}" | head -n 1)
    LBUFFER="${LBUFFER}${line}${RBUFFER}"
    zle reset-prompt
    CURSOR=$(( cur + point ))
}

add-zsh-hook precmd _bayesh_update
add-zle-hook-widget zle-line-init bayesh_infer_cmd
export BAYESH_AVOID_IF_EMPTY