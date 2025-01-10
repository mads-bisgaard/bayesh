from pathlib import Path
from typing import Final
import json
import os
from ._db import create_db

_BAYESH_DIR_ENV_VAR: Final[str] = "BAYESH_DIR"


class BayeshSettings(dict):
    def __init__(self):
        super().__init__()
        self["bayesh_dir"] = os.environ.get(
            _BAYESH_DIR_ENV_VAR, f"{Path.home() / '.bayesh'}"
        )
        self.bayesh_dir.resolve()
        self.bayesh_dir.mkdir(parents=True, exist_ok=True)

        self["db"] = f"{self.bayesh_dir / 'bayesh.db'}"
        if not self.db.is_file():
            create_db(self.db)

        self.process_commands = True

    @property
    def bayesh_dir(self) -> Path:
        return Path(self["bayesh_dir"])

    @property
    def db(self) -> Path:
        return Path(self["db"])
