from pathlib import Path
from bayesh.cli import record_event
from bayesh._db import Row, get_row
from faker import Faker
from .test_db import get_n_rows


def test_record_event(db: Path, row: Row, faker: Faker):
    _row = get_row(db, Path(row.cwd), row.previous_cmd, row.current_cmd)
    assert _row is None
    record_event(Path(row.cwd), row.previous_cmd, row.current_cmd)
    _row = get_row(db, Path(row.cwd), row.previous_cmd, row.current_cmd)
    assert _row is not None
    assert _row.event_counter == 1
    record_event(Path(row.cwd), row.previous_cmd, row.current_cmd)
    _row = get_row(db, Path(row.cwd), row.previous_cmd, row.current_cmd)
    assert _row is not None
    assert _row.event_counter == 2
    _row = get_row(db, Path(row.cwd), row.previous_cmd, row.current_cmd)
    record_event(Path(row.cwd), row.previous_cmd, faker.text())
    assert get_n_rows(db=db) == 2
