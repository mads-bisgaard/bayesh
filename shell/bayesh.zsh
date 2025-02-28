#!/usr/bin/env zsh

autoload -Uz add-zsh-hook
# add hook for updating bayesh db
# shellcheck source=shell/bayesh.sh
source "$(dirname $0)"/bayesh.sh
add-zsh-hook precmd _bayesh_update

if [[ ! -n "$TMUX" ]]; then

    function bayesh_infer_cmd() {
        local cur
        local result
        local point
        
        cur="${CURSOR}"
        result=$(_bayesh_infer_cmd)
        line=$(echo "${result}" | tail -n 1);point=$(echo "${result}" | head -n 1)
        LBUFFER="${LBUFFER}${line}${RBUFFER}"
        zle reset-prompt
        CURSOR=$(( cur + point ))
    }
    zle -N infer_cmd bayesh_infer_cmd
    bindkey '^E' infer_cmd # Ctrl-e    
    return

fi

if ! fzf-tmux-server 1> /dev/null; then echo "Error: fzf-tmux-server doesn't seem to be functional." >&2; return; fi

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
bindkey '^E' start_or_kill_server


function zle-line-init() {
    if [[ -n "$BAYESH_SERVER_CONFIG" ]]; then
        fifo=$(mktemp -u)
        mkfifo "$fifo"
        (
            bayesh infer-cmd "$(pwd)" "${BAYESH_CMD}" > "$fifo" &
            echo "change-query()" | fzf-tmux-server post -c "$BAYESH_SERVER_CONFIG" 2> /dev/null &
            echo "reload(cat $fifo; rm $fifo)" | fzf-tmux-server post -c "$BAYESH_SERVER_CONFIG" 2> /dev/null &
        )
    fi
}
zle -N zle-line-init


function zle-line-pre-redraw() {
    if [[ -n "$BAYESH_SERVER_CONFIG" ]]; then
        ( echo "change-query("$BUFFER")" | fzf-tmux-server post -c "$BAYESH_SERVER_CONFIG" 2> /dev/null & )
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
bindkey '^[^M' select # Alt-Enter

function bayesh_up() {
    if [[ -n "$BAYESH_SERVER_CONFIG" ]]; then
        ( echo "up" | fzf-tmux-server post -c "$BAYESH_SERVER_CONFIG" 2> /dev/null & )
    fi    
}
zle -N up bayesh_up
bindkey '^[[1;3A' up # Alt-uparrow

function bayesh_down() {
    if [[ -n "$BAYESH_SERVER_CONFIG" ]]; then
        ( echo "down" | fzf-tmux-server post -c "$BAYESH_SERVER_CONFIG" 2> /dev/null & )
    fi    
}
zle -N down bayesh_down
bindkey '^[[1;3B' down # Alt-downarrow

trap 'fzf-tmux-server kill -c "$BAYESH_SERVER_CONFIG" &> /dev/null' EXIT
