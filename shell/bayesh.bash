#!/bin/bash

function bayesh_update() {
    local cmd
    local last_hist

    cmd=$(fc -ln -1 | awk '{$1=$1};1')
    last_hist=$(history | tail -1 | md5sum)

    if [[ "${last_hist}" == "${BAYESH_LAST_HIST}" ]]; then
        return
    fi    

    ( bayesh record-event "${BAYESH_PWD}" "${BAYESH_CMD}" "${cmd}" ) & disown


    BAYESH_PWD=$(pwd)
    export BAYESH_PWD
    BAYESH_CMD=${cmd}
    export BAYESH_CMD
    BAYESH_LAST_HIST=${last_hist}
    export BAYESH_LAST_HIST
}

function bayesh_infer_cmd() {
    ( 
    local inferred_cmds

    inferred_cmds=$(bayesh infer-cmd "$(pwd)" "${BAYESH_CMD}")
    hist=$(fc -ln -200 | awk '{$1=$1};1')
    fzf --scheme=history \
        --exact \
        --no-sort \
        --bind="start:reload(echo '${inferred_cmds}')" \
        --bind="zero:reload(echo '${hist}'; echo '{q}')"
    )
}

BAYESH_PWD=$(pwd)
export BAYESH_PWD
BAYESH_CMD=""
export BAYESH_CMD
BAYESH_LAST_HIST=""
export BAYESH_LAST_HIST