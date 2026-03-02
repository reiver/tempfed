package librefs

import (
	gourl "net/url"
)

// gozaar (noun): one who carries out / performs / executes
const PathPrefix string = "/gozaar/"

func Actor(host string, actor string) string {
	var url = gourl.URL{
		Scheme: "https",
		Host:   host,
//@TODO: make the path join safer.
		Path:   PathPrefix + actor,
	}

	return url.String()
}

func ActorInBox(host string, actor string) string {
//@TODO: make the path join safer.
	return Actor(host, actor) + "/inbox"
}

func ActorOutBox(host string, actor string) string {
//@TODO: make the path join safer.
	return Actor(host, actor) + "/outbox"
}

func SharedInBox(host string) string {
	var url = gourl.URL{
		Scheme: "https",
		Host:   host,
		Path:   "/inbox",
	}

	return url.String()
}
