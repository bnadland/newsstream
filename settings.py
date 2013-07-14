DEBUG = True

# Configuration
APPNAME = 'newsstream'
ADMINS = ['benjamin.nadland@freenet.de']
SQLALCHEMY_DATABASE_URI = 'postgresql://newsstream:newsstream@localhost/newsstream'

# DERIVED SETTINGS
ASSETS_DEBUG = False
SQLALCHEMY_ECHO = DEBUG
