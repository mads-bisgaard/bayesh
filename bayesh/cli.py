from typer import Typer
from pathlib import Path
from datetime import datetime
from ._db import get_row, update_row, insert_row, Row
from ._settings import BayeshSettings

cli = Typer()


@cli.command()
def record_event(cwd: Path, previous_cmd: str, current_cmd):
    db = BayeshSettings().db
    if row := get_row(
        db=db, cwd=cwd, previous_cmd=previous_cmd, current_cmd=current_cmd
    ):
        count = row.event_counter + 1
        last_modified = datetime.now()
        update_row(db=db, row=row, event_counter=count, last_modified=last_modified)
    else:
        row = Row(
            cwd=f"{cwd.resolve()}",
            previous_cmd=previous_cmd,
            current_cmd=current_cmd,
            event_counter=1,
            last_modified=datetime.now(),
        )
        insert_row(db=db, row=row)
