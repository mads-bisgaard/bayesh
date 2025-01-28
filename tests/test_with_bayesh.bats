#!/bin/bash
# tests must be run from the root directory of the repo

setup_file() {

    # install and expose bayesh on PATH
    venv="$(mktemp -d)/.venv"
    # shellcheck source=/dev/null
    python -m venv "${venv}" && source "${venv}/bin/activate"
    python -m pip install .
    ln -s "${venv}/bin/bayesh" "/usr/local/bin/bayesh"
    bayesh --help
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

@test "check only record new commands" {
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
    run inotifywait --event modify --timeout 1 "${db}"
    [ "$status" -eq 0 ]

    run bash -c "sqlite3 ${db} 'select count(*) from events'"
    [ "$status" -eq 0 ]
    assert_output '2'
    run bash -c "sqlite3 ${db} 'select event_counter from events'"
    [ "$status" -eq 0 ]
    assert_output '1
1'
}
