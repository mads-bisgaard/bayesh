from pathlib import Path
from typing import Final
import os
from ._db import create_db

_BAYESH_DIR_ENV_VAR: Final[str] = "BAYESH_DIR"


class BayeshSettings(dict):
    def __init__(self):
        super().__init__()
        _bayesh_dir = Path(
            os.environ.get(_BAYESH_DIR_ENV_VAR, f"{Path.home() / '.bayesh'}")
        )
        _bayesh_dir.resolve()
        _bayesh_dir.mkdir(parents=True, exist_ok=True)
        self[_BAYESH_DIR_ENV_VAR] = f"{_bayesh_dir}"

        self["db"] = f"{self.bayesh_dir / 'bayesh.db'}"
        if not self.db.is_file():
            create_db(self.db)

    @property
    def bayesh_dir(self) -> Path:
        return Path(self[_BAYESH_DIR_ENV_VAR])

    @property
    def db(self) -> Path:
        return Path(self["db"])
