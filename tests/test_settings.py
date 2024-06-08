
import pytest
from bayesh._settings import _BAYESH_DIR_ENV_VAR, BayeshSettings
from pathlib import Path
import pydantic

@pytest.mark.parametrize("env_var_set", [True, False])
def test_settings(monkeypatch, env_var_set: bool, tmp_path: Path):
    if env_var_set:
        monkeypatch.setenv(_BAYESH_DIR_ENV_VAR, f"{tmp_path.resolve()}")
        _ = BayeshSettings()
    else:
        with pytest.raises(pydantic.ValidationError):
            BayeshSettings()
    