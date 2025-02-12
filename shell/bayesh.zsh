#!/usr/bin/env zsh

# shellcheck source=shell/bayesh.sh
source "$(dirname $0)"/bayesh.sh

function bayesh_infer_cmd() {
    token_regex="<STRING>|<PATH>"

    fifo=$(_bayesh_infer_cmd)
    while true; do
        line=$(tail -1 < "$fifo")
        position="${#line}"
        if echo "${line}" | grep -boq -E "${token_regex}"; then
            position=$(echo "${line}" | grep -bo -E "${token_regex}" | cut -d: -f1 | head -n1)
        fi
        prompt=$(echo "${line}" | sed -E "s/(${token_regex})//g")

        LBUFFER="${LBUFFER}${prompt}"
        zle reset-prompt
        export CURSOR="${position}"
    done    
}

add-zsh-hook precmd _bayesh_update
export BAYESH_AVOID_IF_EMPTY
