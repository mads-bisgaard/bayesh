import shlex
from pathlib import Path
from enum import StrEnum


class Tokens(StrEnum):
    PATH = "<PATH>"
    STRING = "<STRING>"


def ansi_color_tokens(cmds: str) -> str:
    cmds = cmds.replace(Tokens.PATH, f"\033[94m{Tokens.PATH}\033[0m")
    cmds = cmds.replace(Tokens.STRING, f"\033[94m{Tokens.STRING}\033[0m")
    return cmds


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
            cmd = cmd.replace(p, Tokens.PATH)
        elif " " in p and not Path(p).exists():
            cmd = cmd.replace(p, Tokens.STRING)

    return cmd
