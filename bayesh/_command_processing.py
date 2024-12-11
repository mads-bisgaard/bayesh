import shlex
from pathlib import Path


def process_cmd(cmd: str) -> str:
    parser = shlex.shlex(cmd, posix=True, punctuation_chars=True)
    parser.whitespace_split = True
    parts = list(parser)
    for p in parts:
        ii = parts.index(p)  # *first* index
        if " " in p:
            cmd = cmd.replace(f'"{p}"', "<MSG>")
        elif (
            Path(p).exists()
            and ii > 0
            and not parts[ii - 1].endswith(
                ("(", ")", ";", "<", ">", "|", "&")
            )  # https://docs.python.org/3/library/shlex.html#improved-compatibility-with-shells
        ):  # allow paths in 0th position: pointing to executable
            cmd = cmd.replace(p, "<PATH>")

    return cmd
