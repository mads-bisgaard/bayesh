import click
from pathlib import Path
from datetime import datetime
import json
from ._db import get_row, update_row, insert_row, Row, infer_current_cmd
from ._settings import BayeshSettings
from ._command_processing import process_cmd, ansi_color_tokens


@click.group()
def cli():
    pass


@cli.command("record-event")
@click.argument("cwd", type=click.Path(path_type=Path))
@click.argument("previous_cmd")
@click.argument("current_cmd")
def record_event(cwd: Path, previous_cmd: str, current_cmd: str):
    settings = BayeshSettings()
    previous_cmd = process_cmd(previous_cmd)
    current_cmd = process_cmd(current_cmd)

    if row := get_row(
        db=settings.db, cwd=cwd, previous_cmd=previous_cmd, current_cmd=current_cmd
    ):
        count = row.event_counter + 1
        last_modified = datetime.now()
        update_row(
            db=settings.db, row=row, event_counter=count, last_modified=last_modified
        )
    else:
        row = Row(
            cwd=f"{cwd.resolve()}",
            previous_cmd=previous_cmd,
            current_cmd=current_cmd,
            event_counter=1,
            last_modified=datetime.now(),
        )
        insert_row(db=settings.db, row=row)


@cli.command("infer-cmd")
@click.argument("cwd", type=click.Path(path_type=Path))
@click.argument("previous_cmd")
def infer_cmd(cwd: Path, previous_cmd: str):
    settings = BayeshSettings()
    previous_cmd = process_cmd(previous_cmd)

    results = infer_current_cmd(db=settings.db, cwd=cwd, previous_cmd=previous_cmd)
    click.echo(ansi_color_tokens("\n".join(results)), color=True)


@cli.command("print-settings")
def print_settings():
    click.echo(json.dumps(BayeshSettings()))
