from pathlib import Path
from typing import Final
from enum import StrEnum
from datetime import datetime
from typing import NamedTuple
from sqlalchemy import create_engine, Column, Integer, String, DateTime, Index
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker
from sqlalchemy.dialects.sqlite import insert

_TABLE: Final[str] = "events"
Base = declarative_base()


class Columns(StrEnum):
    cwd = "cwd"
    previous_cmd = "previous_cmd"
    current_cmd = "current_cmd"
    event_counter = "event_counter"
    last_modified = "last_modified"


class Event(Base):
    __tablename__ = _TABLE
    cwd = Column(String, primary_key=True)
    previous_cmd = Column(String, primary_key=True)
    current_cmd = Column(String, primary_key=True)
    event_counter = Column(Integer, nullable=False)
    last_modified = Column(DateTime, nullable=False)
    __table_args__ = (Index("idx_event_counter", "event_counter"),)


class Row(NamedTuple):
    cwd: Path | str
    previous_cmd: str
    current_cmd: str
    event_counter: int
    last_modified: datetime


def create_db(db: Path) -> None:
    engine = create_engine(f"sqlite:///{db}")
    Base.metadata.create_all(engine)


def insert_row(db: Path, row: Row) -> None:
    engine = create_engine(f"sqlite:///{db}")
    Session = sessionmaker(bind=engine)
    session = Session()
    event = Event(
        cwd=str(row.cwd),
        previous_cmd=row.previous_cmd,
        current_cmd=row.current_cmd,
        event_counter=row.event_counter,
        last_modified=row.last_modified,
    )
    session.add(event)
    session.commit()
    session.close()


def update_row(db: Path, row: Row, event_counter: int, last_modified: datetime) -> None:
    engine = create_engine(f"sqlite:///{db}")
    Session = sessionmaker(bind=engine)
    session = Session()
    event = (
        session.query(Event)
        .filter_by(
            cwd=str(row.cwd), previous_cmd=row.previous_cmd, current_cmd=row.current_cmd
        )
        .first()
    )
    if event:
        event.event_counter = event_counter
        event.last_modified = last_modified
        session.commit()
    session.close()


def get_row(db: Path, cwd: Path, previous_cmd: str, current_cmd: str) -> Row | None:
    engine = create_engine(f"sqlite:///{db}")
    Session = sessionmaker(bind=engine)
    session = Session()
    event = (
        session.query(Event)
        .filter_by(cwd=str(cwd), previous_cmd=previous_cmd, current_cmd=current_cmd)
        .first()
    )
    session.close()
    return (
        None
        if event is None
        else Row(
            cwd=Path(event.cwd),
            previous_cmd=event.previous_cmd,
            current_cmd=event.current_cmd,
            event_counter=event.event_counter,
            last_modified=event.last_modified,
        )
    )


def infer_current_cmd(db: Path, cwd: Path, previous_cmd: str) -> list[str]:
    engine = create_engine(f"sqlite:///{db}")
    Session = sessionmaker(bind=engine)
    session = Session()
    events = (
        session.query(Event.current_cmd)
        .filter_by(cwd=str(cwd), previous_cmd=previous_cmd)
        .order_by(Event.event_counter.desc())
        .all()
    )
    session.close()
    return [event.current_cmd for event in events]
