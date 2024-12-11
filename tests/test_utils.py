from bayesh._command_processing import sanitize_cmd
from pathlib import Path
from typing import Final, NamedTuple
import csv
import pytest
from parse import parse
import os

_COMMANDS_FILE: Final[Path] = (
    Path(__file__).parent / "data" / "sanitized_bash_commands.csv"
)
assert _COMMANDS_FILE.is_file()


class CommandPair(NamedTuple):
    raw_cmd: str
    sanitized_cmd: str


def _get_commands() -> list[tuple[str, str]]:
    with open(_COMMANDS_FILE, mode="r") as f:
        csv_reader = csv.reader(f, escapechar="\\")
        return [CommandPair(*row) for row in csv_reader]


@pytest.mark.parametrize(
    "cmdpair", _get_commands(), ids=lambda x: _get_commands().index(x)
)
def test_sanitize_cmd(tmp_path: Path, cmdpair: CommandPair):
    sanitized_cmd = cmdpair.sanitized_cmd
    if "<PATH>" in sanitized_cmd:
        sanitized_cmd = sanitized_cmd.replace("<PATH>", "{}")
        parsed_path = parse(sanitized_cmd, cmdpair.raw_cmd)
        if parsed_path is not None:
            os.chdir(tmp_path)
            for p in parsed_path:
                Path(p).mkdir(parents=True, exist_ok=True)
    assert sanitize_cmd(cmdpair.raw_cmd) == cmdpair.sanitized_cmd
