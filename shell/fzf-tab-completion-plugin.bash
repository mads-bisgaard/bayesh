

function bayesh_autocomplete() {
    local input
    input=$(echo "$READLINE_LINE" | awk '{$1=$1};1')
    if [[ "$input" == "" ]]; then
        local result
        result=$(bayesh_infer_cmd)
        READLINE_LINE="${result}"
        READLINE_POINT=${#result}        
    else
        fzf_bash_completion
    fi
}
