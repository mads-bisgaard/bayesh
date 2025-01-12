from pathlib import Path
from bayesh.cli import record_event
from bayesh._db import Row, get_row
from .test_db import get_n_rows
from click.testing import CliRunner


def test_record_event(db: Path, row: Row):
    runner = CliRunner()

    _row = get_row(db, Path(row.cwd), row.previous_cmd, row.current_cmd)
    assert _row is None
    runner.invoke(record_event, [row.cwd, row.previous_cmd, row.current_cmd])
    _row = get_row(db, Path(row.cwd), row.previous_cmd, row.current_cmd)
    assert _row is not None
    assert _row.event_counter == 1
    runner.invoke(record_event, [row.cwd, row.previous_cmd, row.current_cmd])
    _row = get_row(db, Path(row.cwd), row.previous_cmd, row.current_cmd)
    assert _row is not None
    assert _row.event_counter == 2
    _row = get_row(db, Path(row.cwd), row.previous_cmd, row.current_cmd)
    _random_cmd = row.current_cmd + "something else"
    runner.invoke(record_event, [row.cwd, row.previous_cmd, _random_cmd])
    assert get_n_rows(db=db) == 2
