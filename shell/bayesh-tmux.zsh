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
            unset BAYESH_SERVER_CONFIG
            return 
        fi
    fi
    BAYESH_SERVER_CONFIG=$(fzf-tmux-server start)
    export BAYESH_SERVER_CONFIG
}
zle -N start_or_kill_server bayesh_start_or_kill_server
bindkey '^s' start_or_kill_server


function zle-line-init() {
    if _bayesh_config; then
        ( 
            echo "change-query("$BUFFER")" | fzf-tmux-server post -c "$BAYESH_SERVER_CONFIG" &
            echo "reload(bayesh infer-cmd \""$(pwd)"\" \""${BAYESH_CMD}"\")" | fzf-tmux-server post -c "$BAYESH_SERVER_CONFIG" &
        )
    fi
}
zle -N zle-line-init


function zle-line-pre-redraw() {
    if _bayesh_config; then
        ( 
            echo "change-query("$BUFFER")" | fzf-tmux-server post -c "$BAYESH_SERVER_CONFIG" &
        )
    fi
}
zle -N zle-line-pre-redraw


function bayesh_select() {
    local token_regex
    token_regex="<STRING>|<PATH>"

    if _bayesh_config; then
        cmd=$(fzf-tmux-server get -c "$BAYESH_SERVER_CONFIG" 2> /dev/null | jq -r .current.text)
        p="${#cmd}"
        if echo "${cmd}" | grep -boq -E "${token_regex}"; then
            p=$(echo "${cmd}" | grep -bo -E "${token_regex}" | cut -d: -f1 | head -n1)
        fi
        BUFFER=$(echo "${cmd}" | sed -E "s/(${token_regex})//g")
        zle -R
        CURSOR="$p"
    fi    
}
zle -N select bayesh_select
bindkey '^[[1;5C' select # Ctrl-rightarrow

function bayesh_up() {
    if _bayesh_config; then
        (
            echo "up" | fzf-tmux-server post -c "$BAYESH_SERVER_CONFIG" 2> /dev/null &
        )
    fi    
}
zle -N up bayesh_up
bindkey '^[[1;5A' up # Ctrl-uparrow

function bayesh_down() {
    if _bayesh_config; then
        (
            echo "down" | fzf-tmux-server post -c "$BAYESH_SERVER_CONFIG" 2> /dev/null &
        )
    fi    
}
zle -N down bayesh_down
bindkey '^[[1;5B' down # Ctrl-downarrow

# TODO: this trap still doesn't work as expected
# trap "fzf-tmux-server kill -c \"${BAYESH_SERVER_CONFIG}\"" EXIT HUP INT QUIT TERM
