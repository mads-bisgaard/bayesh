from bayesh._utils import sanitize_cmd
from pathlib import Path
from typing import Final, NamedTuple
import csv
import pytest

_COMMANDS_FILE: Final[Path] = (
    Path(__file__).parent / "data" / "sanitized_bash_commands.csv"
)
assert _COMMANDS_FILE.is_file()


class CommandPair(NamedTuple):
    raw_cmd: str
    sanitized_cmd: str


def _get_sanitized_commands() -> list[tuple[str, str]]:
    with open(_COMMANDS_FILE, mode="r") as f:
        csv_reader = csv.reader(f)
        return [CommandPair(*row) for row in csv_reader]


@pytest.mark.parametrize("cmdpair", _get_sanitized_commands(), ids=lambda x: x.raw_cmd)
def test_sanitize_cmd(cmdpair: CommandPair):
    assert sanitize_cmd(cmdpair.raw_cmd) == cmdpair.sanitized_cmd
