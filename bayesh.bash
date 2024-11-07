#!/bin/bash

#set -e
#set -u

export BAYESH_DIR="${HOME}/Development/bayesh/tmp"

function bayesh_update() {
    local cmd
    local histcmd

    cmd=$(fc -ln -1 | xargs)
    histcmd="${HISTCMD}"
    if [[ "${histcmd}" -eq "${BAYESH_HISTCMD}" ]]; then
        return
    fi    
    ( bayesh record-event "${BAYESH_PWD}" "${BAYESH_CMD}" "${cmd}" ) & disown


    BAYESH_PWD=$(pwd)
    export BAYESH_PWD
    BAYESH_CMD=${cmd}
    export BAYESH_CMD
    BAYESH_HISTCMD=${histcmd}
    export BAYESH_HISTCMD
}

if [[ -n "$PROMPT_COMMAND" ]]; then
    PROMPT_COMMAND="$PROMPT_COMMAND; bayesh_update"
else
    PROMPT_COMMAND='bayesh_update'
fi

BAYESH_PWD=$(pwd)
export BAYESH_PWD
BAYESH_CMD=""
export BAYESH_CMD
BAYESH_HISTCMD="-1"
export BAYESH_HISTCMD