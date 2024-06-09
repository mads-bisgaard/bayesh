
import pytest
from bayesh._settings import _BAYESH_DIR_ENV_VAR, BayeshSettings
from pathlib import Path
import pydantic

@pytest.mark.parametrize("tmp_bayesh_dir_exists", [True, False])
def test_bayesh_dir(tmp_bayesh_dir: Path):
    dir_exists = tmp_bayesh_dir.is_dir()
    _ = BayeshSettings()
    if not dir_exists:
        assert tmp_bayesh_dir.is_dir()

def test_bayesh_dir_file_path(monkeypatch, tmp_path: Path):
    tmp_file = tmp_path / "myfile.txt"
    tmp_file.write_text("hi there")
    assert tmp_file.is_file()
    monkeypatch.setenv(_BAYESH_DIR_ENV_VAR, f"{tmp_file.resolve()}")
    with pytest.raises(FileExistsError):
        _ = BayeshSettings()
    
    