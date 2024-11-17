#!/bin/bash

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

function bayesh_infer_cmd() {
    ( 
    local inferred_cmds

    inferred_cmds=$(bayesh infer-cmd "$(pwd)" "${BAYESH_CMD}")
    fzf --scheme=history --no-sort \
    --bind="start:reload(echo '${inferred_cmds}')" \
    --bind="zero:reload(echo '${inferred_cmds}'; echo '{q}')" \
    --bind="one:reload(echo '${inferred_cmds}'; echo '{q}')"
    )
}

BAYESH_PWD=$(pwd)
export BAYESH_PWD
BAYESH_CMD=""
export BAYESH_CMD
BAYESH_HISTCMD="-1"
export BAYESH_HISTCMD