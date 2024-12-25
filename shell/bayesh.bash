#!/bin/bash

function bayesh_post_process_command() {
    local processed_cmd
    local tokens
    local read_point_str
    local -n result_array
    result_array="$1"
    processed_cmd="$2"

    if ! echo "${processed_cmd}" | grep -q '<[A-Z]*>' ;then
        result_array+=(${#processed_cmd})
        result_array+=("${processed_cmd}")
        return
    fi

    tokens=$(echo "${processed_cmd}" | grep -o '<[A-Z]*>')
    read_point_str="${processed_cmd%%"$(echo "${tokens}" | head -n 1)"*}"
    for substr in ${tokens}; do
        processed_cmd="${processed_cmd//${substr}/}"
    done

    result_array+=(${#read_point_str})
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
    local chosen_cmd
    local result
    
    pwd
    echo "${BAYESH_CMD}"

    chosen_cmd=$( 
    local inferred_cmds
    inferred_cmds=$(bayesh infer-cmd "$(pwd)" "${BAYESH_CMD}")
    fzf --scheme=history \
        --exact \
        --no-sort \
        --bind="start:reload(echo '${inferred_cmds}')" \
        --bind="zero:reload(echo '${inferred_cmds}'; echo '{q}')"
    )

    result=()
    bayesh_post_process_command result "${chosen_cmd}"
    READLINE_LINE="${READLINE_LINE:0:${READLINE_POINT}}${result[1]}${READLINE_LINE:${READLINE_POINT}}"
    READLINE_POINT=$(("${READLINE_POINT}" + result[0]))
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
