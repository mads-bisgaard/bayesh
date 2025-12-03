#!/usr/bin/env bash
# tests must be run from the root directory of the repo
bats_require_minimum_version 1.5.0
repo=$(pwd)

setup_file() {
    # install bayesh binary
    [ -d "${repo}/build" ] || exit 1
    [ -f "${repo}/build/bayesh" ] || exit 1
    cp "${repo}/build/bayesh" /usr/local/bin/bayesh
}

setup() {
    load /batslib/bats-support/load
    load /batslib/bats-assert/load
    BAYESH_DIR=$(mktemp -d)
    export BAYESH_DIR
}

teardown() {
    rm -rf "${BAYESH_DIR}"
}

@test "test bayesh installed" {
    run bayesh --help
    [ "$status" -eq 0 ]
}

@test "test source script" {
    run bash -c \
    '
    source < (bayesh --bash)
    [[ -v BAYESH_PWD ]] && [[ -v BAYESH_CMD ]] && [[ -v BAYESH_LAST_HIST ]]
    '
    [ "$status" -eq 0 ]
}

@test "test only record new command" {
    #shellcheck source=./shell/bayesh.bash
    source <(bayesh --bash)
    command="random command ${RANDOM}"
    db=$(bayesh settings | jq -r .BAYESH_DB)
    
    run bash -c "sqlite3 ${db} 'select count(*) from events'"
    [ "$status" -eq 0 ]
    assert_output '0'
    
    # simulate run command
    history -s "${command}"
    # wait for insertion into db (https://linux.die.net/man/1/inotifywait)
    inotifywait --event modify --timeout 5 "${db}" &
    monitor_pid=$!
    _bayesh_update
    [ "$status" -eq 0 ]
    wait "$monitor_pid"

    run bash -c "sqlite3 ${db} 'select count(*) from events'"
    [ "$status" -eq 0 ]
    assert_output '1'
    run bash -c "sqlite3 ${db} 'select event_counter from events'"
    [ "$status" -eq 0 ]
    assert_output '1'

    # simulate simply pressing 'enter' with no command
    _bayesh_update && sleep 1

    run bash -c "sqlite3 ${db} 'select count(*) from events'"
    [ "$status" -eq 0 ]
    assert_output '1'
    run bash -c "sqlite3 ${db} 'select event_counter from events'"
    [ "$status" -eq 0 ]
    assert_output '1'    

    # simulate running new command
    inotifywait --event modify --timeout 5 "${db}" &
    monitor_pid=$!
    history -s "${command} ${RANDOM}"
    _bayesh_update 
    wait "$monitor_pid"

    run bash -c "sqlite3 ${db} 'select count(*) from events'"
    [ "$status" -eq 0 ]
    assert_output '2'
    run bash -c "sqlite3 ${db} 'select event_counter from events'"
    [ "$status" -eq 0 ]
    assert_output '1
1'
}

@test "test inference function (no tokens)" {
    #shellcheck source=./shell/bayesh.bash
    bayesh --bash | source

    db=$(bayesh settings | jq -r .BAYESH_DB)

    cwd=$(mktemp -d)
    previous_cmd="previous command ${RANDOM}"
    current_cmd="current command ${RANDOM}"

    run bash -c "sqlite3 ${db} \"insert into events (cwd, previous_cmd, current_cmd, event_counter, last_modified) values ('${cwd}', '${previous_cmd}', '${current_cmd}', 1, '$(date)')\""
    [ "$status" -eq 0 ]

    run bash -c "sqlite3 ${db} 'select count(*) from events'"
    [ "$status" -eq 0 ]
    assert_output '1'

    # setup
    cd "${cwd}" 
    export BAYESH_CMD="${previous_cmd}"
    export FZF_DEFAULT_OPTS="--filter=\"${current_cmd}\"" 
    run _bayesh_infer_cmd </dev/null
    [ "$status" -eq 0 ]
    assert_output "${#current_cmd}
${current_cmd}"

}
