
from bayesh._settings import BayeshSettings
from bayesh._db import create_db
import pytest


def test_db_creation(tmp_bayesh_dir):
    settings = BayeshSettings()
    create_db(settings.db)