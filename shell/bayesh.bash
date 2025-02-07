#!/usr/bin/env bash

# shellcheck source=shell/bayesh.sh
source "$(dirname "${BASH_SOURCE[0]}")"/bayesh.sh

function bayesh_infer_cmd() {
    local result
    local line
    local point
    
    result=$(_bayesh_infer_cmd)
    line=$(echo "${result}" | tail -n 1);point=$(echo "${result}" | head -n 1)
    READLINE_LINE="${READLINE_LINE:0:${READLINE_POINT}}${line}${READLINE_LINE:${READLINE_POINT}}"
    READLINE_POINT=$(("${READLINE_POINT}" + point))
}

export -f _bayesh_post_process_command

if [[ -z "$PROMPT_COMMAND" ]]; then
    PROMPT_COMMAND="_bayesh_update;"
else
    PROMPT_COMMAND="${PROMPT_COMMAND%;}; _bayesh_update;"
fi
export PROMPT_COMMAND
