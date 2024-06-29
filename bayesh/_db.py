import sqlite3
from pathlib import Path
from typing import Final
from enum import StrEnum
from pydantic import PositiveInt, BaseModel
from datetime import datetime
from typing import NamedTuple

_TABLE: Final[str] = "events"


class Columns(StrEnum):
    cwd = "cwd"
    previous_cmd = "previous_cmd"
    current_cmd = "current_cmd"
    event_counter = "event_counter"
    last_modified = "last_modified"


class Row(NamedTuple):
    cwd: Path | str
    previous_cmd: str
    current_cmd: str
    event_counter: PositiveInt
    last_modified: datetime


def create_db(db: Path) -> None:
    create_query = f"""
    CREATE TABLE {_TABLE} (
        {Columns.cwd} TEXT,
        {Columns.previous_cmd} TEXT,
        {Columns.current_cmd} TEXT,
        {Columns.event_counter} INTEGER CHECK({Columns.event_counter} >= 0),
        {Columns.last_modified} DATETIME,
        PRIMARY KEY ({Columns.cwd}, {Columns.previous_cmd}, {Columns.current_cmd})
    )
    """
    index_query = (
        f"CREATE INDEX idx_event_counter ON {_TABLE} ({Columns.event_counter})"
    )
    with sqlite3.connect(f"{db}") as conn:
        cursor = conn.cursor()
        cursor.execute(create_query)
        cursor.execute(index_query)
        conn.commit()


def insert_row(db: Path, row: Row) -> None:
    assert db.is_file()  # nosec
    insert_statement = f"""
    INSERT INTO {_TABLE}({Columns.cwd},{Columns.previous_cmd},{Columns.current_cmd},{Columns.event_counter},{Columns.last_modified})
    VALUES(?,?,?,?,?) 
    """
    with sqlite3.connect(f"{db}") as conn:
        cursor = conn.cursor()
        cursor.execute(insert_statement, row)
        conn.commit()


def get_row(db: Path, cwd: Path, previous_cmd: str, current_cmd) -> Row | None:
    assert db.is_file()  # nosec
    get_statement = f"""
    SELECT * FROM {_TABLE}
    WHERE {Columns.cwd} = ? AND {Columns.previous_cmd} = ? AND {Columns.current_cmd} = ?
    """
    params = (f"{cwd}", previous_cmd, current_cmd)
    with sqlite3.connect(f"{db}") as conn:
        cursor = conn.cursor()
        cursor.execute(get_statement, params)
        result = cursor.fetchone()
        return None if result is None else Row(*result)


def update_row(db: Path, row: Row, event_counter: PositiveInt, last_modified: datetime):
    assert db.is_file()  # nosec
    update_statement = f"""
    UPDATE {_TABLE}
    SET {Columns.event_counter} = ?, {Columns.last_modified} = ?
    WHERE {Columns.cwd} = ? AND {Columns.previous_cmd} = ? AND {Columns.current_cmd} = ?
    """
    params = (
        event_counter,
        last_modified,
        f"{row.cwd}",
        row.previous_cmd,
        row.current_cmd,
    )
    with sqlite3.connect(f"{db}") as conn:
        cursor = conn.cursor()
        cursor.execute(update_statement, params)
        conn.commit()
