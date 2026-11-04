#!/usr/bin/env zsh

# Native completion runs in a process forked from the user's shell, so that it
# sees the user's completions, styles, aliases and functions, and so that a slow
# completion function can't block the prompt.

zmodload zsh/zutil

# Replaces compadd in the forked process: prints each match as a complete
# command line instead of adding it to the list of matches.
function _bayesh_compadd() {
    # these options are used for querying matches, not for adding them
    if [[ ${@[1,(i)(-|--)]} == *-(O|A|D)\ * ]]; then
        builtin compadd "$@"
        return $?
    fi

    # the options can refer to array parameters by name, so nothing else may be
    # declared before calling compadd
    local -a __bayesh_hits
    builtin compadd -A __bayesh_hits "$@"
    (( $#__bayesh_hits )) || return 0
    # The completion system retries with the next tag, matcher and completer until
    # matches have been added, which would print them again
    builtin compadd "$@"

    setopt localoptions norcexpandparam extendedglob
    local -A apre hpre hsuf asuf
    zparseopts -E P:=apre p:=hpre S:=asuf s:=hsuf

    # -f means that the matches are files. Don't be fooled by -default-
    local -i dirsuf=0
    if [[ -z $hsuf && -z $asuf && "${${@//-default-/}% -# *}" == *-[[:alnum:]]#f* ]]; then
        dirsuf=1
    fi

    local hit dsuf
    local pre=${LBUFFER%${(b)IPREFIX}${(b)PREFIX}}
    local post=${RBUFFER#${(b)SUFFIX}${(b)ISUFFIX}}
    for hit in "${__bayesh_hits[@]}"; do
        (( dirsuf )) && [[ -d $hit ]] && dsuf=/ || dsuf=
        print -r -- "$pre$QIPREFIX$IPREFIX$apre$hpre$hit$dsuf$hsuf$asuf$ISUFFIX$QISUFFIX$post"
    done
    return 0
}

function _bayesh_capture() {
    exec < /dev/null > "$_BAYESH_FIFO" 2> /dev/null
    if (( ! ${+_comps} )); then
        autoload -Uz compinit
        compinit -u -D
    fi
    function compadd() { _bayesh_compadd "$@"; }
    _main_complete
}

# Writes the completions of the current command line to $_BAYESH_FIFO
function _bayesh_capture_widget() {
    _bayesh_capture &!
    _BAYESH_WRITER_PID=$!
    compstate[insert]=''
    compstate[list]=''
}
zle -C _bayesh_capture_completions complete-word _bayesh_capture_widget
