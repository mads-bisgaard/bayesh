#!/usr/bin/env bash

response_pipe=$1
if [[ ! -w $response_pipe ]]; then
    echo "invalid pipe" 
    exit 1
fi

function run_fzf() {
    # shellcheck disable=SC2016
    fzf --listen \
        --bind 'start:execute-silent(echo -e "BAYESH_PORT=$FZF_PORT\nBAYESH_PANE_ID=$TMUX_PANE" > '"${response_pipe}"' & )' \
        --scheme=history \
        --no-sort \
        --exact \
        --ansi \
        --border=none \
        --info=inline-right \
        --layout=reverse \
        --margin=0 \
        --padding=0 \
        --no-mouse \
        --no-info
}

run_fzf "${response_pipe}"