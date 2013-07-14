#!./env/bin/python
import subprocess
from flask.ext.script import Manager
from sqlalchemy import create_engine

import settings
import scraper
from app import app
from models import Base, Link

manager = Manager(app)

@manager.command
def createdb():
    subprocess.call(['sudo', '-u', 'postgres', 'dropdb', 'newsstream'])
    subprocess.call(['sudo', '-u', 'postgres', 'createdb', '-O', 'newsstream', '-E', 'utf8', 'newsstream'])
    engine = create_engine(app.config['SQLALCHEMY_DATABASE_URI'], echo=True)
    Base.metadata.create_all(bind=engine)

@manager.command
def scrape():
    scraper.scrape()

@manager.command
def update():
    scraper.update()

if __name__ == '__main__':
    manager.run()
