
from bayesh._settings import BayeshSettings
from bayesh._db import create_db, insert_row, _TABLE, Columns, Row, get_row, update_row
import pytest
from pathlib import Path
import sqlite3
from faker import Faker
from datetime import datetime

def _get_n_rows(db: Path) -> int:
    assert db.is_file()
    with sqlite3.connect(f"{db}") as conn:
        cursor = conn.cursor()
        cursor.execute(f"SELECT COUNT(*) FROM {_TABLE}")
        return cursor.fetchone()[0]

def test_db_creation(tmp_path: Path):
    db_file  = tmp_path / "my.db"
    create_db(db_file)
    assert db_file.is_file()

def test_insert_row(tmp_path: Path, db: Path, faker: Faker):
    assert db.is_file()
    assert _get_n_rows(db) == 0

    previous_cmd = faker.text()
    current_cmd = faker.text()
    event_counter = faker.random_int(min=1, max=1000)
    insert_row(db, Row(f"{tmp_path}", previous_cmd, current_cmd, event_counter, datetime.now()))
    assert _get_n_rows(db) == 1

    with sqlite3.connect(db) as conn:
        cursor = conn.cursor()
        cursor.execute(f"SELECT * FROM {_TABLE}")
        results = cursor.fetchall()
        assert len(results) == 1
        row = Row(*results[0])
        row.cwd == tmp_path
        row.previous_cmd == previous_cmd
        row.current_cmd == current_cmd
        row.event_counter == event_counter

def test_db_unique_key(tmp_path: Path, db: Path, faker: Faker):
    assert db.is_file()
    assert _get_n_rows(db) == 0

    previous_cmd = faker.text()
    current_cmd = faker.text()
    event_counter = faker.random_int(min=1, max=1000)
    insert_row(db, Row(f"{tmp_path}", previous_cmd, current_cmd, event_counter, datetime.now()))
    assert _get_n_rows(db) == 1
    with pytest.raises(sqlite3.IntegrityError):
        insert_row(db, Row(f"{tmp_path}", previous_cmd, current_cmd, faker.random_int(min=1, max=1000), datetime.now()))

def test_get_row(db: Path, faker: Faker, tmp_path: Path, row: Row):
    assert get_row(db, tmp_path, faker.text(), faker.text()) == None
    insert_row(db, row)
    _row = get_row(db, row.cwd, row.previous_cmd, row.current_cmd)
    assert _row is not None
    assert row.cwd == _row.cwd
    assert row.previous_cmd == _row.previous_cmd
    assert row.current_cmd == _row.current_cmd
    assert row.event_counter == _row.event_counter

def test_update_row(db: Path, faker: Faker, row: Row):
    assert _get_n_rows(db) == 0
    insert_row(db, row)

    _event_counter = faker.random_int(min=1, max=1000)
    _last_modified = datetime.now()
    update_row(db, row, _event_counter, last_modified=_last_modified)
    _row = get_row(db, row.cwd, row.previous_cmd, row.current_cmd)  
    assert _row.event_counter == _event_counter
    assert _row.last_modified == f"{_last_modified}"