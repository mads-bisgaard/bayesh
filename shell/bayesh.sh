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
    inferred_cmds=$(bayesh infer-cmd "$(pwd)" "${BAYESH_CMD}")
    
    if [ "${BAYESH_AVOID_IF_EMPTY+set}" ] && [ -z "$(echo "${inferred_cmds}" | awk '{$1=$1};1')" ]; then
        exit 1
    fi
    
    fifo=$(mktemp -u)
    mkfifo "$fifo"
    (
    echo "${inferred_cmds}" | fzf \
        --scheme=history \
        --no-sort \
        --exact \
        --bind="zero:abort" \
        --bind="tab:execute-silent(echo {} > ${fifo})" \
        --bind="focus:execute-silent(echo {} > ${fifo})" \
        --ansi \
        --border=none \
        --preview-window=border-rounded,up:1:wrap \
        --header-first \
        --info=inline-right \
        --layout=reverse \
        --margin=0 \
        --padding=0 \
        --height=30% \
        --no-mouse &
    )

    echo "$fifo"

    while true; do
        line=$(tail -1 < "$fifo")
        position="${#line}"
        if echo "${line}" | grep -boq -E "${token_regex}"; then
            position=$(echo "${line}" | grep -bo -E "${token_regex}" | cut -d: -f1 | head -n1)
        fi
        prompt=$(echo "${line}" | sed -E "s/(${token_regex})//g")

        LBUFFER="${LBUFFER}${prompt}"
        zle reset-prompt
        export CURSOR="${position}"
    done

}

BAYESH_PWD=$(pwd)
export BAYESH_PWD
BAYESH_CMD=""
export BAYESH_CMD
BAYESH_LAST_HIST=""
export BAYESH_LAST_HIST
