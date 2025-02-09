#!/bin/sh

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
    
    token_regex="<STRING>|<PATH>"
    chosen_cmd=$( 
    inferred_cmds=$(bayesh infer-cmd "$(pwd)" "${BAYESH_CMD}")
    
    if [ "${BAYESH_AVOID_IF_EMPTY+set}" ] && [ -z "$(echo "${inferred_cmds}" | awk '{$1=$1};1')" ]; then
        exit 1
    fi
    echo "${inferred_cmds}" | fzf \
        --scheme=history \
        --no-sort \
        --exact \
        --bind="zero:print-query" \
        --bind="ctrl-q:print-query" \
        --ansi \
        --preview="echo {} | sed -E 's/(${token_regex})//g'" \
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

    position="${#chosen_cmd}"
    if echo "${chosen_cmd}" | grep -boq -E "${token_regex}"; then
        position=$(echo "${chosen_cmd}" | grep -bo -E "${token_regex}" | cut -d: -f1 | head -n1)
    fi
    echo "${position}"
    echo "${chosen_cmd}" | sed -E "s/(${token_regex})//g"
}

BAYESH_PWD=$(pwd)
export BAYESH_PWD
BAYESH_CMD=""
export BAYESH_CMD
BAYESH_LAST_HIST=""
export BAYESH_LAST_HIST
