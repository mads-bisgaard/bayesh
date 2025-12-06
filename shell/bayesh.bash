#!/usr/bin/env bash


function bayesh_infer_cmd() {
    local result
    local line
    local point

    result=$(_bayesh_infer_cmd)
    line=$(echo "${result}" | tail -n 1);point=$(echo "${result}" | head -n 1)
    READLINE_LINE="${READLINE_LINE:0:${READLINE_POINT}}${line}${READLINE_LINE:${READLINE_POINT}}"
    READLINE_POINT=$(("${READLINE_POINT}" + point))
}

if [[ -z "$PROMPT_COMMAND" ]]; then
    PROMPT_COMMAND="_bayesh_update;"
else
    PROMPT_COMMAND="${PROMPT_COMMAND%;}; _bayesh_update;"
fi
export PROMPT_COMMAND

if [[ $- == *i* ]]; then
    bind -x '"\C-e":"bayesh_infer_cmd"'
fi