#!/usr/bin/env zsh

if ! fzf-tmux-server 1> /dev/null; then echo "Error: fzf-tmux-server doesn't seem to be functional." >&2; return; fi

# add hook for updating bayesh db
# shellcheck source=shell/bayesh.sh
source "$(dirname $0)"/bayesh.sh
add-zsh-hook precmd _bayesh_update


#functions for communicating with server

function _bayesh_config() {
    if [[ -n "$BAYESH_SERVER_CONFIG" ]]; then 
        return 0
    fi
    return 1
}


function bayesh_start_or_kill_server() {

    if _bayesh_config; then
        if fzf-tmux-server get -c "$BAYESH_SERVER_CONFIG" &> /dev/null; then
            fzf-tmux-server kill -c "$BAYESH_SERVER_CONFIG"
            return 
        fi
    fi
    BAYESH_SERVER_CONFIG=$(fzf-tmux-server start)
    export BAYESH_SERVER_CONFIG
}

function zle-line-init() {
    if _bayesh_config; then
        ( 
            if fzf-tmux-server get -c "$BAYESH_SERVER_CONFIG" &> /dev/null; then echo "reload(bayesh infer-cmd \""$(pwd)"\" \""${BAYESH_CMD}"\")" | fzf-tmux-server post -c "$BAYESH_SERVER_CONFIG"; fi &
        )
    fi
}


# TODO: this trap still doesn't work as expected
trap "fzf-tmux-server kill -c \"${BAYESH_SERVER_CONFIG}\"" EXIT HUP INT QUIT TERM

zle -N start_or_kill_server bayesh_start_or_kill_server
zle -N zle-line-init

bindkey '^s' start_or_kill_server
