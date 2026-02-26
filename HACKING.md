# HACKING.md

## Program Overview

**tempfed** is an application written in the Go programming-language (golang) that streams public posts from Fediverse instances and stores data in a ClickHouse database server.

When the Fediverse instance is a Mastodon server or a Mastodon-compatible server (such as `fedi.buzz`) it connects to the server, streams public posts via SSE (sever-sent events), decodes streaming events into ActivityPub Notes, and writes to a ClickHouse database.
