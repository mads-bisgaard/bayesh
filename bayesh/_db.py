
import sqlite3
from pathlib import Path

def create_db(db_path: Path):
    query = """
    CREATE TABLE events (
        cwd TEXT,
        previous_cmd TEXT,
        current_cmd TEXT,
        event_counter INTEGER CHECK(event_counter >= 0),
        last_modified DATETIME,
        PRIMARY KEY (cwd, previous_cmd, current_cmd)
    )
    """
    with sqlite3.connect(f"{db_path}") as conn:
        cursor = conn.cursor()
        cursor.execute(query)
        conn.commit()
