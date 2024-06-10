
from bayesh._settings import BayeshSettings
from bayesh._db import create_db
import pytest
from pathlib import Path

def test_db_creation(tmp_path: Path):
    db_file  = tmp_path / "my.db"
    create_db(db_file)
    assert db_file.is_file()