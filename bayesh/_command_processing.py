import shlex
from pathlib import Path
import re
from typing import Final

_PATH_REGEX_PATTERN: Final[str] = (
    r"(\<PATH\>)\s+(\1)+"  # Matches <PATH> followed by one or more copies of itself
)
_MSG_REGEX_PATTERN: Final[str] = (
    r"(\<MSG\>)\s+(\1)+"  # Matches <MSG> followed by one or more copies of itself
)


def process_cmd(cmd: str) -> str:
    parser = shlex.shlex(cmd, posix=True, punctuation_chars=True)
    parser.whitespace_split = True
    parts = list(parser)
    for ii, p in enumerate(parts):
        if cmd.count(p) > 1:
            continue
        if (
            Path(p).exists()
            and p != "."
            and ii > 0
            and not parts[ii - 1].endswith(
                ("(", ")", ";", "<", ">", "|", "&")
            )  # https://docs.python.org/3/library/shlex.html#improved-compatibility-with-shells
        ):  # allow paths in 0th position: pointing to executable
            cmd = cmd.replace(p, "<PATH>")
        elif " " in p:
            cmd = cmd.replace(f"{p}", "<MSG>")

    return cmd
