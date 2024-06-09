
import pytest
from pathlib import Path
from typing import Iterator

from bayesh._settings import _BAYESH_DIR_ENV_VAR


@pytest.fixture
def tmp_bayesh_dir(tmp_path: Path, monkeypatch, tmp_bayesh_dir_exists: bool) -> Iterator[Path]:
    bayesh_dir = tmp_path / ".bayesh"
    assert not bayesh_dir.is_dir()
    if tmp_bayesh_dir_exists:
        bayesh_dir.mkdir()
    monkeypatch.setenv(_BAYESH_DIR_ENV_VAR, f"{bayesh_dir.resolve()}")
    yield bayesh_dir