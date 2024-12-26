import typer
from pathlib import Path
from datetime import datetime
from ._db import get_row, update_row, insert_row, Row, infer_current_cmd
from ._settings import BayeshSettings
from ._command_processing import process_cmd

cli = typer.Typer()


@cli.command()
def record_event(cwd: Path, previous_cmd: str, current_cmd: str):
    settings = BayeshSettings()
    if settings.process_commands:
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


@cli.command()
def infer_cmd(cwd: Path, previous_cmd: str):
    settings = BayeshSettings()
    if settings.process_commands:
        previous_cmd = process_cmd(previous_cmd)

    results = infer_current_cmd(db=settings.db, cwd=cwd, previous_cmd=previous_cmd)
    typer.echo("\n".join(results))


@cli.command()
def print_settings():
    typer.echo(BayeshSettings().model_dump_json())
