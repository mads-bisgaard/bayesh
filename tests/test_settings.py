import pytest
from bayesh._settings import _BAYESH_DIR_ENV_VAR, BayeshSettings
from pathlib import Path
import shutil


@pytest.mark.parametrize("bayesh_dir_exists", [True, False])
def test_bayesh_dir(tmp_bayesh_dir: Path, bayesh_dir_exists: bool):
    assert tmp_bayesh_dir.is_dir()
    if not bayesh_dir_exists:
        shutil.rmtree(tmp_bayesh_dir)
        assert not tmp_bayesh_dir.is_dir()
    settings = BayeshSettings()
    assert tmp_bayesh_dir.is_dir()
    assert settings.db.is_file()


def test_bayesh_dir_file_path(monkeypatch, tmp_path: Path):
    tmp_file = tmp_path / "myfile.txt"
    tmp_file.write_text("hi there")
    assert tmp_file.is_file()
    monkeypatch.setenv(_BAYESH_DIR_ENV_VAR, f"{tmp_file.resolve()}")
    with pytest.raises(FileExistsError):
        _ = BayeshSettings()
