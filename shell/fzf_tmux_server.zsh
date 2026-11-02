#!/usr/bin/env zsh

zmodload zsh/net/tcp zsh/system

# Starts the fzf-tmux server in a pane to the right of the current one and prints
# its config to stdout. Pass the config to the other functions in order to access
# the server.
_fzf_tmux_server_start() {
    local fifo
    fifo=$(mktemp -u)
    mkfifo "$fifo"
    # The config is written on the first load event, because a reload sent to fzf
    # before it has finished the initial load can get lost.
    # shellcheck disable=SC2016
    script=(
        'echo "" | fzf '
        "--listen "
        '--bind='\''load:unbind(load)+execute-silent(echo "{ \"url\" : \"http://localhost:$FZF_PORT\", \"tmux_pane_id\": \"$TMUX_PANE\" }" > '"$fifo"' & )'\''' 
        "--scheme=history "
        "--no-sort "
        "--no-input "
        "--exact "
        "--ansi "
        "--border=none "
        "--info=inline-right "
        "--layout=reverse "
        "--margin=0 "
        "--padding=0 "
        "--no-mouse "
        "--no-info "
        "--border=rounded"
        "--style=minimal"
        "--no-scrollbar"
    )

    tmux split-window -t "$TMUX_PANE" -h -l 35% -d "${script[*]}"

    cat "$fifo"
    rm "$fifo"
}

# Sends an HTTP request from within the shell and prints the response body.
# The shell talks to the server on every keystroke, so this must not fork.
# Usage: _fzf_tmux_server_request <server config> <method> <path> [body]
_fzf_tmux_server_request() {
    # Content-Length counts bytes, so strings must not be treated as multibyte
    emulate -L zsh
    setopt no_multibyte
    local config="$1" method="$2" request_path="$3" body="$4"
    local port="${${config##*localhost:}%%[^0-9]*}"
    local REPLY fd chunk response="" headers=""

    [[ -n "$port" ]] || return 1
    # fzf rejects a Content-Length of 0
    [[ -z "$body" ]] || headers="Content-Length: ${#body}"$'\r\n'
    ztcp 127.0.0.1 "$port" 2> /dev/null || return 1
    fd="$REPLY"
    print -rn -u "$fd" -- "$method $request_path HTTP/1.1"$'\r\n'"Host: localhost"$'\r\n'"Connection: close"$'\r\n'"$headers"$'\r\n'"$body"
    while sysread -t 1 -i "$fd" chunk; do
        response+="$chunk"
    done
    ztcp -c "$fd"

    [[ "$response" == HTTP/1.1\ 200* ]] || return 1
    print -r -- "${response#*$'\r\n\r\n'}"
}

# Ends fzf, which also closes its pane. This must not depend on starting a
# process, since it also runs while a shell is being killed along with its pane.
# Usage: _fzf_tmux_server_kill <server config>
_fzf_tmux_server_kill() {
    _fzf_tmux_server_post "$1" abort
}

# Prints the state of the server, without the list of matches.
# Usage: _fzf_tmux_server_get <server config>
_fzf_tmux_server_get() {
    _fzf_tmux_server_request "$1" GET "/?limit=1"
}

# Usage: _fzf_tmux_server_post <server config> <fzf action>
_fzf_tmux_server_post() {
    _fzf_tmux_server_request "$1" POST / "$2" > /dev/null
}
