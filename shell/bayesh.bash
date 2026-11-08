#!/usr/bin/env bash

_bayesh_update() {

    cmd=$(fc -ln -1 | awk '{$1=$1};1')
    last_hist=$(history | tail -1 | md5sum)

    if [ "${last_hist}" = "${BAYESH_LAST_HIST}" ]; then
        return
    fi

    ( bayesh record-event "${BAYESH_PWD}" "${BAYESH_CMD}" "${cmd}" & )

    BAYESH_PWD=$(pwd)
    export BAYESH_PWD
    BAYESH_CMD=${cmd}
    export BAYESH_CMD
    BAYESH_LAST_HIST=${last_hist}
    export BAYESH_LAST_HIST
}

BAYESH_PWD=$(pwd)
export BAYESH_PWD
BAYESH_CMD=""
export BAYESH_CMD
BAYESH_LAST_HIST=""
export BAYESH_LAST_HIST


function bayesh_infer_cmd() {
    local result
    local line
    local point

    result=$(_bayesh_infer_cmd "${BAYESH_CMD}")
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