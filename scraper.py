import feedparser
import datetime
import time

from praw import Reddit
from sqlalchemy import create_engine, Column, Integer, String, or_
from sqlalchemy.orm import sessionmaker
from sqlalchemy.sql import exists

import settings
from sources import SOURCES
from models import Link

def reddit(subreddit):
    result = []
     
    r = Reddit('newsstream v0.2 by /u/bnadland')

    for i in r.get_subreddit(subreddit).get_hot():
        result.append({
            'title': i.title,
            'url': i.url
        })

    return result

def feed(feedurl, summary = False):
    result = []

    feed = feedparser.parse(feedurl)

    if not feed.entries:
        return result

    for i in feed.entries:
        # title
        title = i.title
        if summary:
            title = i.summary

        # url
        url = i.get('feedburner_origlink', i.link)

        # published
        published = datetime.datetime.now()
        if i.get('updated_parsed', None):
            published = datetime.datetime.fromtimestamp(time.mktime(i.updated_parsed))
        elif i.get('published_parsed', None):
            published = datetime.datetime.fromtimestamp(time.mktime(i.published_parsed))

        result.append({
            'title': title,
            'url': url,
            'published': published,
        })       

    return result

def scrape():
    db = create_engine(settings.SQLALCHEMY_DATABASE_URI, echo=True)
    Session = sessionmaker(bind=db)
    dbsession = Session()
    
    link_collection = set([x[0] for x in dbsession.query(Link.url).all()])

    for source in SOURCES.keys():
        links = []

        # decide on scraper
        scraper = SOURCES[source].get('scraper', 'feed')
        url = SOURCES[source].get('url', '')

        if scraper == 'feed':
            links = feed(url)
        elif scraper == 'summary':
            links = feed(url, summary = True)
        elif scraper == 'reddit':
            links = reddit(source)

        # save new entries 
        if not links:
            continue

        for i in links:
            if i['url'] not in link_collection:
                link_collection.add(i['url'])
                link = Link()
                link.title     = i['title']
                link.url       = i['url']
                link.source    = source
                link.rank      = SOURCES[source].get('rank', '1')
                link.type      = SOURCES[source].get('type', 'politics')
                link.published = i.get('published', datetime.datetime.now())
                dbsession.add(link)
        dbsession.commit()

def update():
    db = create_engine(settings.SQLALCHEMY_DATABASE_URI, echo=True)
    Session = sessionmaker(bind=db)
    dbsession = Session()

    for source in SOURCES:
        dbsession.query(Link) \
            .filter(Link.source == source) \
            .update({
                    'rank': SOURCES[source].get('rank', '1'),
                    'type': SOURCES[source].get('type', 'politics'),
                },
                synchronize_session=False
            )
    dbsession.commit()
