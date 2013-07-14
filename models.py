#!./env/bin/python
import datetime
from sqlalchemy import Column, Integer, String, DateTime

from database import Base

class Link(Base):
    __tablename__ = 'link'

    id = Column(Integer, primary_key=True)
    title = Column(String)
    url = Column(String)
    source = Column(String)
    type = Column(String, default="politics")
    rank = Column(Integer, default=1)
    published = Column(DateTime, default=datetime.datetime.now)
    created = Column(DateTime, default=datetime.datetime.now)
