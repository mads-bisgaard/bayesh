from pydantic_settings import BaseSettings
from pathlib import Path
from pydantic import model_validator, Field
from typing import Final
from typing_extensions import Self
from ._db import create_db

_BAYESH_DIR_ENV_VAR: Final[str] = "BAYESH_DIR"


class BayeshSettings(BaseSettings):
    bayesh_dir: Path = Field(Path.home() / ".bayesh", alias=_BAYESH_DIR_ENV_VAR)
    process_commands: bool = Field(True, alias="BAYESH_PROCESS_COMMANDS")

    @model_validator(mode="after")
    def check_dir(self) -> Self:
        self.bayesh_dir.resolve()
        self.bayesh_dir.mkdir(parents=True, exist_ok=True)
        self.bayesh_dir.resolve()
        if not self.db.is_file():
            create_db(self.db)
        return self

    @property
    def db(self) -> Path:
        return self.bayesh_dir / "bayesh.db"
