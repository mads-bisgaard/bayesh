
import pytest
from pathlib import Path
from typing import Iterator
from datetime import datetime
from bayesh._settings import _BAYESH_DIR_ENV_VAR, BayeshSettings
from bayesh._db import Row

@pytest.fixture
def tmp_bayesh_dir(tmp_path: Path, monkeypatch) -> Iterator[Path]:
    bayesh_dir = tmp_path / "bayesh"
    assert not bayesh_dir.is_dir()
    bayesh_dir.mkdir()
    monkeypatch.setenv(_BAYESH_DIR_ENV_VAR, f"{bayesh_dir.resolve()}")
    yield bayesh_dir

@pytest.fixture
def db(tmp_bayesh_dir) -> Iterator[Path]:
    yield BayeshSettings().db

@pytest.fixture
def row(tmp_path: Path, faker: Path) -> Iterator[Row]:
    yield Row(
        cwd=f"{tmp_path}",
        previous_cmd=faker.text(),
        current_cmd=faker.text(),
        event_counter=faker.random_int(min=1, max=10000),
        last_modified=datetime.now()
    )