import re
from typing import Final
from bashlex import ast, parsesingle

_PATH_REGEX: Final[str] = r"^(/[^/\0]+)+$"


def _reconstruct_word_node(node: ast.node) -> str:
    """process bashlex node of kind 'word'"""
    assert hasattr(node, "kind") and node.kind == "word"  # nosec

    if " " in node.word:
        return f'"{node.word}"'
    return node.word


def _reconstruct_cmd_from_ast(cmd_ast: ast.node) -> str:
    def walk_and_collect(node: ast.node):
        parts = []
        if hasattr(node, "parts") and len(node.parts) > 0:
            for part in node.parts:
                parts += walk_and_collect(part)
        elif hasattr(node, "kind") and node.kind == "operator":
            parts.append(node.op)
        elif hasattr(node, "kind") and node.kind == "pipe":
            parts.append(node.pipe)
        elif hasattr(node, "kind") and node.kind == "word":
            parts.append(_reconstruct_word_node(node))
        return parts

    return " ".join(walk_and_collect(cmd_ast))


def sanitize_cmd(cmd: str) -> str:
    parts = parsesingle(cmd)
    return cmd
