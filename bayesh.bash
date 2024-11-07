#!/bin/bash

#set -e
#set -u

export BAYESH_DIR="${HOME}/Development/bayesh/tmp"

function bayesh_update() {
    local cmd
    cmd=$(fc -ln -1 | xargs)
    bayesh record-event "${BAYESH_PWD}" "${BAYESH_CMD}" "${cmd}"


    BAYESH_PWD=$(pwd)
    export BAYESH_PWD
    BAYESH_CMD=${cmd}
    export BAYESH_CMD
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
