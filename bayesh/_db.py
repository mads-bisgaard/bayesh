
import sqlite3
from pathlib import Path
from typing import Final
from enum import StrEnum
from ._settings import BayeshSettings

_TABLE: Final[str] = "events"

class Columns(StrEnum):
    cwd = "cwd"
    previous_cmd = "previous_cmd"
    current_cmd = "current_cmd"
    event_counter = "event_counter"
    last_modified = "last_modified"



def create_db(db_path: Path):
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
    index_query = f"CREATE INDEX idx_event_counter ON {_TABLE} ({Columns.event_counter})"
    with sqlite3.connect(f"{db_path}") as conn:
        cursor = conn.cursor()
        cursor.execute(create_query)
        cursor.execute(index_query)
        conn.commit()


def upsert(cwd: Path, previous_cmd: str, current_cmd: str):
    _db = BayeshSettings().db
    if not _db.is_file():
        create_db(_db)
    