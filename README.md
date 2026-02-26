# tempfed

**tempfed** is a _temporary caching server_ for the Fediverse and Social Web.

## Motivation

Traditionally, the most common cause for why a Fediverse or Social Web server crashed is
—
its _storage drive_ fills up with data from other servers (typically for caching purposes).
This data from other servers includes:
posts,
media,
profile data,
avatar images,
header images,
etc.

**tempfed** tries to address part of this problem
—
by providing a service (to Fediverse and Social Web servers) that stores **posts** for a _short_ period of time.

That way Fediverse and Social Web servers do _not_ have to store this data themselves.

Here are some example of who Fediverse and Social Web servers could use **tempfed**:

* a social-media application could get the post data for a user's home-feed from **tempfed**
* a custom feed service might get the post data from **tempfed**
* a social-media application that allows its user to search posts might get the post data from **tempfed**

## Author

Software **tmpfed** was written by [Charles Iliya Krempeaux](http://reiver.link)
