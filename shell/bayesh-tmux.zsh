#!/usr/bin/env zsh

if ! fzf-tmux-server 1> /dev/null; then echo "Error: fzf-tmux-server doesn't seem to be functional." >&2; return; fi

# add hook for updating bayesh db
# shellcheck source=shell/bayesh.sh
source "$(dirname $0)"/bayesh.sh
add-zsh-hook precmd _bayesh_update


#functions for communicating with server

function bayesh_start_or_kill_server() {

    if [[ -n "$BAYESH_SERVER_CONFIG" ]]; then
        if fzf-tmux-server get -c "$BAYESH_SERVER_CONFIG" &> /dev/null; then
            fzf-tmux-server kill -c "$BAYESH_SERVER_CONFIG"
            unset BAYESH_SERVER_CONFIG
            return 
        fi
    fi
    BAYESH_SERVER_CONFIG=$(fzf-tmux-server start)
    export BAYESH_SERVER_CONFIG
    zle-line-init
}
zle -N start_or_kill_server bayesh_start_or_kill_server
bindkey '^s' start_or_kill_server


function zle-line-init() {
    if [[ -n "$BAYESH_SERVER_CONFIG" ]]; then
        fifo=$(mktemp -u)
        mkfifo "$fifo"
        (
            bayesh infer-cmd "$(pwd)" "${BAYESH_CMD}" > "$fifo" &
            echo "change-query()" | fzf-tmux-server post -c "$BAYESH_SERVER_CONFIG" &
            echo "reload(cat $fifo; rm $fifo)" | fzf-tmux-server post -c "$BAYESH_SERVER_CONFIG" &
        )
    fi
}
zle -N zle-line-init


function zle-line-pre-redraw() {
    if [[ -n "$BAYESH_SERVER_CONFIG" ]]; then
        ( 
            echo "change-query("$BUFFER")" | fzf-tmux-server post -c "$BAYESH_SERVER_CONFIG" &
        )
    fi
}
zle -N zle-line-pre-redraw


function bayesh_select() {
    local token_regex
    local cmd
    local p
    token_regex="<STRING>|<PATH>"

    if [[ -n "$BAYESH_SERVER_CONFIG" ]]; then
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
    if [[ -n "$BAYESH_SERVER_CONFIG" ]]; then
        (
            echo "up" | fzf-tmux-server post -c "$BAYESH_SERVER_CONFIG" 2> /dev/null &
        )
    fi    
}
zle -N up bayesh_up
bindkey '^[[1;5A' up # Ctrl-uparrow

function bayesh_down() {
    if [[ -n "$BAYESH_SERVER_CONFIG" ]]; then
        (
            echo "down" | fzf-tmux-server post -c "$BAYESH_SERVER_CONFIG" 2> /dev/null &
        )
    fi    
}
zle -N down bayesh_down
bindkey '^[[1;5B' down # Ctrl-downarrow

trap 'fzf-tmux-server kill -c "$BAYESH_SERVER_CONFIG" &> /dev/null' EXIT
