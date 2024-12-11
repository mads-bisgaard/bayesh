from bayesh._command_processing import process_cmd
from pathlib import Path
from typing import Final, Iterable
import csv
import pytest
from pydantic import BaseModel
from pytest_mock import mocker, MockerFixture

_COMMANDS_FILE: Final[Path] = Path(__file__).parent / "data" / "processed_bash_commands"
assert _COMMANDS_FILE.is_file()


class CommandPairTestData(BaseModel):
    raw_cmd: str
    sanitized_cmd: str
    required_paths: list[Path] = []


def _get_commands() -> list[tuple[str, str]]:
    examples = []
    with open(_COMMANDS_FILE, mode="r") as f:
        for line in f:
            examples.append(CommandPairTestData.model_validate_json(line))
    return examples


@pytest.fixture
def commands_with_mocked_paths(
    mocker: MockerFixture, cmdpair: CommandPairTestData
) -> Iterable[CommandPairTestData]:

    class DummyPath:
        def __init__(self, s: str):
            self.s = s

        def exists(self) -> bool:
            if Path(self.s) in cmdpair.required_paths:
                return True
            else:
                return False

    mocker.patch("bayesh._command_processing.Path", DummyPath)
    yield cmdpair


@pytest.mark.parametrize(
    "cmdpair", _get_commands(), ids=lambda x: _get_commands().index(x)
)
def test_sanitize_cmd(commands_with_mocked_paths: CommandPairTestData):
    assert (
        process_cmd(commands_with_mocked_paths.raw_cmd)
        == commands_with_mocked_paths.sanitized_cmd
    )


if __name__ == "__main__":
    # function for generating ndjson test data
    with open(_COMMANDS_FILE, mode="r") as f:
        csv_reader = csv.reader(f, escapechar="\\")
        # print([CommandPair(*row) for row in csv_reader]
        for row in csv_reader:
            print(
                CommandPairTestData(
                    raw_cmd=row[0], sanitized_cmd=row[1]
                ).model_dump_json()
            )
