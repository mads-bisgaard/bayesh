
import sqlite3
from pathlib import Path

def create_db(db_path: Path):
    create_query = """
    CREATE TABLE events (
        cwd TEXT,
        previous_cmd TEXT,
        current_cmd TEXT,
        event_counter INTEGER CHECK(event_counter >= 0),
        last_modified DATETIME,
        PRIMARY KEY (cwd, previous_cmd, current_cmd)
    )
    """
    index_query = "CREATE INDEX idx_event_counter ON events (event_counter)"
    with sqlite3.connect(f"{db_path}") as conn:
        cursor = conn.cursor()
        cursor.execute(create_query)
        cursor.execute(index_query)
        conn.commit()
