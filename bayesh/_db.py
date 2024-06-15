
import sqlite3
from pathlib import Path
from typing import Final
from enum import StrEnum
from pydantic import PositiveInt
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
    cwd: Path
    previous_cmd: str
    current_cmd: str
    event_counter: PositiveInt
    last_modified: datetime



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


def insert_row(db_path: Path, cwd: Path, previous_cmd:str, current_cmd:str, event_counter: PositiveInt):
    assert db_path.is_file() # nosec
    insert_statement = f'''
    INSERT INTO {_TABLE}({Columns.cwd},{Columns.previous_cmd},{Columns.current_cmd},{Columns.event_counter},{Columns.last_modified})
    VALUES(?,?,?,?,?) 
    '''
    values=(f"{cwd.resolve()}", previous_cmd, current_cmd, f"{event_counter}", f"{datetime.now()}")
    with sqlite3.connect(f"{db_path}") as conn:
        cursor = conn.cursor()
        cursor.execute(insert_statement, values)
        conn.commit()