#!/usr/bin/env zsh

if ! fzf-tmux-server 1> /dev/null; then echo "Error: fzf-tmux-server doesn't seem to be functional." >&2; exit 1; fi

function bayesh_start_or_kill_server() {

    if [[ -n "$BAYESH_SERVER_CONFIG" ]] && fzf-tmux-server get -c "$BAYESH_SERVER_CONFIG" > /dev/null; then
        fzf-tmux-server kill -c "$BAYESH_SERVER_CONFIG"
    else
        BAYESH_SERVER_CONFIG=$(fzf-tmux-server start)
        export BAYESH_SERVER_CONFIG
    fi
}
 

trap "fzf-tmux-server kill -c \"${BAYESH_SERVER_CONFIG}\"" EXIT HUP INT QUIT TERM
