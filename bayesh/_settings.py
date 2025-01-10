from pathlib import Path
from typing import Final
import os
from ._db import create_db

_BAYESH_DIR_ENV_VAR: Final[str] = "BAYESH_DIR"


class BayeshSettings:
    def __init__(self):
        self.bayesh_dir = Path(
            os.environ.get(_BAYESH_DIR_ENV_VAR, f"{Path.home() / '.bayesh'}")
        )
        self.bayesh_dir.resolve()
        self.bayesh_dir.mkdir(parents=True, exist_ok=True)
        self.bayesh_dir.resolve()
        if not self.db.is_file():
            create_db(self.db)

        self.process_commands = True

    @property
    def db(self) -> Path:
        return self.bayesh_dir / "bayesh.db"
