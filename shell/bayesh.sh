#!/bin/sh

_bayesh_post_process_command() {
    processed_cmd="$1"

    if ! echo "${processed_cmd}" | grep -q '<[A-Z]*>' ;then
        echo ${#processed_cmd}
        echo "${processed_cmd}"
        return
    fi

    tokens=$(echo "${processed_cmd}" | grep -o '<[A-Z]*>')
    read_point_str="${processed_cmd%%"$(echo "${tokens}" | head -n 1)"*}"
    for substr in ${tokens}; do
        processed_cmd=$(echo "${processed_cmd}" | sed "s/${substr}//g")
    done

    echo ${#read_point_str}
    echo "${processed_cmd}"
}


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

_bayesh_infer_cmd() {
    
    chosen_cmd=$( 
    inferred_cmds=$(bayesh infer-cmd "$(pwd)" "${BAYESH_CMD}")

    fzf --scheme=history \
        --no-sort \
        --exact \
        --bind="start:reload(echo '${inferred_cmds}')" \
        --bind="zero:print-query" \
        --bind="ctrl-q:print-query" \
        --ansi \
        --preview='_bayesh_post_process_command {} | tail -n 1' \
        --border=none \
        --preview-window=border-rounded,up:1:wrap \
        --header="Press Ctrl+q to select query" \
        --header-first \
        --info=inline-right \
        --layout=reverse \
        --margin=0 \
        --padding=0 \
        --height=30%
    ) || return

    _bayesh_post_process_command "${chosen_cmd}"
}

BAYESH_PWD=$(pwd)
export BAYESH_PWD
BAYESH_CMD=""
export BAYESH_CMD
BAYESH_LAST_HIST=""
export BAYESH_LAST_HIST
