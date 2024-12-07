from bayesh._utils import sanitize_cmd
from pathlib import Path
from typing import Final, Generator
import csv
import pytest

_COMMANDS_FILE: Final[Path] = (
    Path(__file__).parent / "data" / "sanitized_bash_commands.csv"
)
assert _COMMANDS_FILE.is_file()


def _get_sanitized_commands() -> list[tuple[str, str]]:
    with open(_COMMANDS_FILE, mode="r") as f:
        csv_reader = csv.reader(f)
        return [tuple(row) for row in csv_reader]


@pytest.mark.parametrize(
    "cmd_expectedcmd", _get_sanitized_commands(), ids=lambda x: x[0]
)
def test_sanitize_cmd(cmd_expectedcmd: tuple[str, str]):
    assert sanitize_cmd(cmd_expectedcmd[0]) == cmd_expectedcmd[1]


if __name__ == "__main__":
    print(_get_sanitized_commands())
