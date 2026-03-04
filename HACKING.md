# HACKING.md

## Program Overview

**tempfed** is an application written in the Go programming-language (golang) that serves as a special-purpose Fediverse node.
It provides ActivityPub actor endpoints via an HTTP server, that serve actor profiles.
The ActivityPub actor endpoints are discoverable via WebFinger.
**tempfed** can stream public posts from Fediverse instances and store data in a ClickHouse database server.

When the Fediverse instance is a Mastodon server or a Mastodon-compatible server (such as `fedi.buzz`) it connects to the server, streams public posts via SSE (server-sent events), decodes streaming events into ActivityPub Notes, and writes to a ClickHouse database.
