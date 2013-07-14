#!./env/bin/python
import datetime

from flask import Flask, render_template
from flask.ext.assets import Environment, Bundle
from htmlmin import minify
from sqlalchemy import distinct, func

import settings
from database import db_session
from models import Link

# init
app = Flask(__name__)
app.config.from_object('settings')

if not app.debug:
    import logging
    from logging.handlers import SMTPHandler
    mail_handler = SMTPHandler('127.0.0.1',
                               'newsstream@wellenmann.de',
                               settings.ADMINS, 'newsstream error')
    mail_handler.setLevel(logging.ERROR)
    app.logger.addHandler(mail_handler)

@app.teardown_request
def shutdown_session(exception=None):
    db_session.remove()

# assets
assets = Environment(app)
assets.url = app.static_url_path

js = Bundle(
    'vendor/js/jquery-1.9.1.js',
    'vendor/js/foundation/foundation.js',
    'vendor/js/foundation/foundation.alerts.js',
    'vendor/js/foundation/foundation.clearing.js',
    'vendor/js/foundation/foundation.cookie.js',
    'vendor/js/foundation/foundation.dropdown.js',
    'vendor/js/foundation/foundation.forms.js',
    'vendor/js/foundation/foundation.joyride.js',
    'vendor/js/foundation/foundation.magellan.js',
    'vendor/js/foundation/foundation.orbit.js',
    'vendor/js/foundation/foundation.placeholder.js',
    'vendor/js/foundation/foundation.reveal.js',
    'vendor/js/foundation/foundation.section.js',
    'vendor/js/foundation/foundation.tooltips.js',
    'vendor/js/foundation/foundation.topbar.js',
    'js/main.js',
    filters='jsmin',
    output='assets/newsstream.js')
assets.register('newsstream.js', js)

modernizr = Bundle(
    'vendor/js/custom.modernizr.js',
    output='assets/modernizr.js')
assets.register('modernizr.js', modernizr)

css = Bundle(
    'vendor/scss/normalize.scss',
    'vendor/scss/foundation.scss',
    'scss/*.scss',
    filters='pyscss, cssmin',
    output='assets/newsstream.css')
assets.register('newsstream.css', css)

# helpers
def get_links(atdate):
    links = db_session.query(Link) \
        .filter(Link.published > "{}".format(atdate)) \
        .filter(Link.published < "{}".format(atdate + datetime.timedelta(days=1))) \
        .order_by(Link.created.asc()) \
        .all()
    
    types = db_session.query(func.count(Link.type).label('count'), Link.type.label('name')) \
        .filter(Link.published > "{}".format(atdate)) \
        .filter(Link.published < "{}".format(atdate + datetime.timedelta(days=1))) \
        .group_by(Link.type) \
        .order_by('count DESC') \
        .all()

    return links, types 

# routes
@app.errorhandler(404)
def error404(error):
    return minify(render_template('404.html')), 404

@app.route('/')
def newsstream():
    today = datetime.date.today()
    links, types = get_links(today)
    return minify(render_template('newsstream.html', links=links, today=today, types=types))

@app.route('/imprint/')
def imprint():
    return minify(render_template('imprint.html'))
