#!/bin/bash

function bayesh_post_process_command() {
    local processed_cmd
    local -n result_array
    result_array="$1"
    processed_cmd="$2"

    result_array+=("${processed_cmd}")

}


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
    local result
    
    result=$( 
    local inferred_cmds

    inferred_cmds=$(bayesh infer-cmd "$(pwd)" "${BAYESH_CMD}")
    fzf --scheme=history \
        --exact \
        --no-sort \
        --bind="start:reload(echo '${inferred_cmds}')" \
        --bind="zero:reload(echo '${inferred_cmds}'; echo '{q}')"
    )

    READLINE_LINE="${result}"
    READLINE_POINT=${#result}    
}

BAYESH_PWD=$(pwd)
export BAYESH_PWD
BAYESH_CMD=""
export BAYESH_CMD
BAYESH_LAST_HIST=""
export BAYESH_LAST_HIST

if [[ -z "$PROMPT_COMMAND" ]]; then
    PROMPT_COMMAND="bayesh_update;"
else
    PROMPT_COMMAND="${PROMPT_COMMAND%;}; bayesh_update;"
fi
export PROMPT_COMMAND
