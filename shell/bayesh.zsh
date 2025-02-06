#!/usr/bin/env zsh

# shellcheck source=shell/bayesh.sh
source "$(dirname $0)"/bayesh.sh

function bayesh_infer_cmd() {
    result=$(_bayesh_infer_cmd)
    line=$(echo "${result}" | tail -n 1);point=$(echo "${result}" | head -n 1)
    LBUFFER="${LBUFFER}${line}${RBUFFER}"
    CURSOR=$(("${CURSOR}" + point))
    CURSOR=$(("${CURSOR}" - "1")) # CURSOR is 1-based in zsh
    zle reset-prompt
}

add-zsh-hook precmd _bayesh_update
add-zle-hook-widget zle-line-init bayesh_infer_cmd
export BAYESH_AVOID_IF_EMPTY