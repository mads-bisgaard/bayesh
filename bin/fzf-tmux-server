#!/usr/bin/env bash


_fzf_tmux_server_start_help() {
    echo "Usage: _fzf_tmux_server_start"
    echo "Starts fzf-tmux server and prints server config to stdout."
    echo "Capture the config and pass it to other methods in order to access the server."
}

_fzf_tmux_server_start() {
    if [[ "$1" == "--help" ]]; then
        _fzf_tmux_server_start_help
        return
    fi
    local fifo
    fifo=$(mktemp -u)
    mkfifo "$fifo"
    # shellcheck disable=SC2016
    script=(
        'echo "" | fzf '
        "--listen "
        '--bind='\''start:execute-silent(echo "{ \"url\" : \"http://localhost:$FZF_PORT\", \"tmux_pane_id\": \"$TMUX_PANE\" }" > '"$fifo"' & )'\''' 
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
    )

    tmux split-window -l 5 -d "${script[*]}"

    jq -c . < "$fifo"
    rm "$fifo"
}


_fzf_tmux_server_parse_config() {
    local input=""
    
    while getopts ":c:" opt; do
        case $opt in
            c)
                input="$OPTARG"
                ;;
            \?)
                echo "Invalid option: -$OPTARG" >&2
                return 1
                ;;
            :)
                echo "Option -$OPTARG requires an argument." >&2
                return 1
                ;;
        esac
    done

    if [[ -z "$input" ]]; then
        echo "Error: config option (-c) is required." >&2
        return 1
    fi

    echo "$input"
}

_fzf_tmux_server_kill_help() {
    echo "Usage: _fzf_tmux_server_kill -c <server config>"
    echo "Kills the server specified by the config."
}

_fzf_tmux_server_kill() {
    if [[ "$1" == "--help" ]]; then
        _fzf_tmux_server_kill_help
        return
    fi
    local config
    local url
    local pane_id
    config=$(_fzf_tmux_server_parse_config "$@") || { _fzf_tmux_server_kill_help; exit 1; }
    url=$(echo "$config" | jq -r .url)
    pane_id=$(echo "$config" | jq -r .tmux_pane_id)
    
    if curl -XGET "$url" &> /dev/null; then
        if ! tmux kill-pane -t "$pane_id"; then
            echo "Could not kill fzf-tmux server" >&2
            exit 1
        fi
    fi
}

_fzf_tmux_server_get_help() {
    echo "Usage: _fzf_tmux_server_get -c <server config>"
    echo "Get the state of the fzf-tmux server."
}

_fzf_tmux_server_get() {
    if [[ "$1" == "--help" ]]; then
        _fzf_tmux_server_get_help
        return
    fi
    local config
    local url
    config=$(_fzf_tmux_server_parse_config "$@") || { _fzf_tmux_server_get_help; exit 1; }
    url=$(echo "$config" | jq -r .url)
    if ! curl -XGET "$url"; then
        echo "GET request failed" >&2
        exit 1
    fi
}

_fzf_tmux_server_post_help() {
    echo "Usage: _fzf_tmux_server_post -c <server config>"
    echo "Post an action to the fzf-tmux-server by piping the request body to stdin"
    # shellcheck disable=SC2016
    echo 'Example: echo "reload(find $(pwd))" | fzf-tmux-server post -c "$config"'
}

_fzf_tmux_server_post() {
    if [[ "$1" == "--help" ]]; then
        _fzf_tmux_server_post_help
        return
    fi
    local config
    local url
    config=$(_fzf_tmux_server_parse_config "$@") || { _fzf_tmux_server_post_help; exit 1; }
    url=$(echo "$config" | jq -r .url)
    if ! curl -XPOST "$url" -d "$(cat)"; then
        echo "POST request failed" >&2
        exit 1
    fi
}

