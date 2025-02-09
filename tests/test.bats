#!/usr/bin/env bash
# tests must be run from the root directory of the repo
bats_require_minimum_version 1.5.0
_repo=$(pwd)
_repo_copy=$(mktemp -d)

setup_file() {
    cp -r "${_repo}" "${_repo_copy}" || exit 1
    cd "${_repo_copy}" || exit 1
    cd "$(ls)" || exit 1
    run ./install.sh
    [ "$status" -eq 0 ] 
}

teardown_file() {
    cd "${_repo}" || exit 1
    rm -rf "${_repo_copy}"
    rm -f /usr/local/bin/bayesh
    run -127 bayesh --help
    [ "$status" -eq 127 ] 
}

setup() {
    bats_load_library bats-support
    bats_load_library bats-assert
    BAYESH_DIR=$(mktemp -d)
    export BAYESH_DIR
}

teardown() {
    rm -rf "${BAYESH_DIR}"
}

@test "test bayesh installed succcessfully" {
    run bayesh --help
    [ "$status" -eq 0 ]
}

@test "test source script" {
    run bash -c \
    '
    source shell/bayesh.bash
    [[ -v BAYESH_PWD ]] && [[ -v BAYESH_CMD ]] && [[ -v BAYESH_LAST_HIST ]]
    '
    [ "$status" -eq 0 ]
}

@test "test only record new command" {
    source shell/bayesh.bash
    command="random command ${RANDOM}"
    db=$(bayesh print-settings | jq -r .db)
    
    run bash -c "sqlite3 ${db} 'select count(*) from events'"
    [ "$status" -eq 0 ]
    assert_output '0'
    
    # simulate run command
    history -s "${command}"
    # wait for insertion into db (https://linux.die.net/man/1/inotifywait)
    _bayesh_update
    run inotifywait --event modify --timeout 5 "${db}"
    [ "$status" -eq 0 ]
    
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
    history -s "${command} ${RANDOM}"
    _bayesh_update 
    run inotifywait --event modify --timeout 5 "${db}"
    [ "$status" -eq 0 ]

    run bash -c "sqlite3 ${db} 'select count(*) from events'"
    [ "$status" -eq 0 ]
    assert_output '2'
    run bash -c "sqlite3 ${db} 'select event_counter from events'"
    [ "$status" -eq 0 ]
    assert_output '1
1'
}

@test "test inference function (no tokens)" {
    source shell/bayesh.bash

    db=$(bayesh print-settings | jq -r .db)

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
    export FZF_DEFAULT_OPTS='--select-1' 
    run _bayesh_infer_cmd
    [ "$status" -eq 0 ]
    assert_output "${#current_cmd}
${current_cmd}"

}
