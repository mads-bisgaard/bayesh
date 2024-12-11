import re
from typing import Final
from bashlex import ast, parsesingle
import shlex
from pathlib import Path

_PATH_REGEX: Final[str] = r"^(/[^/\0]+)+$"


def _reconstruct_word_node(node: ast.node) -> str:
    """process bashlex node of kind 'word'"""
    assert hasattr(node, "kind") and node.kind == "word"  # nosec

    if " " in node.word:
        return f'"{node.word}"'
    return node.word


def _reconstruct_cmd_from_ast(cmd_ast: ast.node) -> str:
    def _walk_and_collect(node: ast.node):
        parts = []

        def _collect_parts(node: ast.node):
            parts = []
            if hasattr(node, "parts") and len(node.parts) > 0:
                for part in node.parts:
                    parts += _walk_and_collect(part)
            return parts

        if hasattr(node, "kind") and node.kind == "operator":
            parts.append(node.op)
        elif hasattr(node, "kind") and node.kind == "pipe":
            parts.append(node.pipe)
        elif hasattr(node, "kind") and node.kind == "pipeline":
            parts += _collect_parts(node)
        elif hasattr(node, "kind") and node.kind == "command":
            parts += _collect_parts(node)
        elif hasattr(node, "kind") and node.kind == "list":
            parts += _collect_parts(node)
        elif hasattr(node, "kind") and node.kind == "commandsubstitution":
            words = _walk_and_collect(node.command)
            if len(words) > 0:
                words[0] = '"$(' + words[0]
                words[-1] = words[-1] + ')"'
            else:
                words = ['"$()"']
            parts += words
        elif hasattr(node, "kind") and node.kind == "assignment":
            words = _collect_parts(node)
            if len(words) > 0:
                words[0] = node.word[: node.word.index("=") + 1] + words[0]
            parts += words
        elif hasattr(node, "kind") and node.kind == "word":
            parts.append(_reconstruct_word_node(node))
        return parts

    return " ".join(_walk_and_collect(cmd_ast))


def sanitize_cmd(cmd: str) -> str:
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
