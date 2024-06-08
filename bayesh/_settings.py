
from pydantic_settings import BaseSettings
from pathlib import Path
from pydantic import field_validator, Field
from typing import Final

_BAYESH_DIR_ENV_VAR: Final[str] = "BAYESH_DIR"


class BayeshSettings(BaseSettings):
    bayesh_dir: Path = Field(alias=_BAYESH_DIR_ENV_VAR)

    @field_validator("bayesh_dir", mode="after")
    def check_dir(cls, v: Path):
        if not v.is_dir():
            raise ValueError("v was not a directory")


    @property
    def db(self) -> Path:
        return self.bayesh_dir / "bayesh.db"