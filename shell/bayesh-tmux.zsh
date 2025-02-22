


function run_fzf_server() {

    set -x # debug
    local fifo
    fifo=$(mktemp -u)
    mkfifo "$fifo"

    tmux split-window -l 4 -d -c "$(realpath $(dirname $0))" "./bayesh-tmux-server.bash $fifo"

    while IFS= read -r line; do
        export "$line"
    done < "$fifo"

    rm "$fifo"
}
 

trap "tmux kill-pane -t \"${BAYESH_PANE_ID}\"" EXIT HUP INT QUIT TERM
